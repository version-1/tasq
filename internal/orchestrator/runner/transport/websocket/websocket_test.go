package websocket

import (
	"bufio"
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestConnectionSendsAndReceivesTextFrames(t *testing.T) {
	t.Parallel()

	server := startMockServer(t, func(t *testing.T, conn net.Conn, reader *bufio.Reader) {
		got := readClientFrame(t, reader)
		if got != `{"id":1}` {
			t.Fatalf("client frame = %q", got)
		}
		writeServerFrame(t, conn, 0x1, []byte(`{"id":1,"result":{}}`))
	})
	defer server.Close()

	conn, err := Dial(context.Background(), "ws://"+server.Addr().String())
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()
	if err := conn.Send(context.Background(), []byte(`{"id":1}`)); err != nil {
		t.Fatalf("send frame: %v", err)
	}
	got, err := conn.Receive(context.Background())
	if err != nil {
		t.Fatalf("receive frame: %v", err)
	}
	if string(got) != `{"id":1,"result":{}}` {
		t.Fatalf("received frame = %q", got)
	}
}

func TestConnectionReturnsEOFOnCloseFrame(t *testing.T) {
	t.Parallel()

	server := startMockServer(t, func(t *testing.T, conn net.Conn, _ *bufio.Reader) {
		writeServerFrame(t, conn, 0x8, nil)
	})
	defer server.Close()

	conn, err := Dial(context.Background(), "ws://"+server.Addr().String())
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()
	_, err = conn.Receive(context.Background())
	if err != io.EOF {
		t.Fatalf("receive error = %v, want EOF", err)
	}
}

func TestConnectionReturnsEOFWhenServerDisconnects(t *testing.T) {
	t.Parallel()

	server := startMockServer(t, func(t *testing.T, conn net.Conn, _ *bufio.Reader) {
		_ = conn.Close()
	})
	defer server.Close()

	conn, err := Dial(context.Background(), "ws://"+server.Addr().String())
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()
	_, err = conn.Receive(context.Background())
	if err != io.EOF {
		t.Fatalf("receive error = %v, want EOF", err)
	}
}

func TestConnectionReceiveHonorsContextDeadline(t *testing.T) {
	t.Parallel()

	server := startMockServer(t, func(t *testing.T, _ net.Conn, _ *bufio.Reader) {
		time.Sleep(200 * time.Millisecond)
	})
	defer server.Close()

	conn, err := Dial(context.Background(), "ws://"+server.Addr().String())
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	_, err = conn.Receive(ctx)
	if err != context.DeadlineExceeded {
		t.Fatalf("receive error = %v, want context deadline exceeded", err)
	}
}

type mockServer struct {
	listener net.Listener
}

func startMockServer(t *testing.T, handler func(*testing.T, net.Conn, *bufio.Reader)) *mockServer {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	server := &mockServer{listener: listener}
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		reader := bufio.NewReader(conn)
		req, err := http.ReadRequest(reader)
		if err != nil {
			t.Errorf("read request: %v", err)
			return
		}
		key := req.Header.Get("Sec-WebSocket-Key")
		if key == "" {
			t.Errorf("missing Sec-WebSocket-Key")
			return
		}
		response := strings.Join([]string{
			"HTTP/1.1 101 Switching Protocols",
			"Upgrade: websocket",
			"Connection: Upgrade",
			"Sec-WebSocket-Accept: " + testAcceptKey(key),
			"",
			"",
		}, "\r\n")
		if _, err := io.WriteString(conn, response); err != nil {
			t.Errorf("write handshake: %v", err)
			return
		}
		handler(t, conn, reader)
	}()
	return server
}

func (s *mockServer) Addr() net.Addr {
	return s.listener.Addr()
}

func (s *mockServer) Close() {
	_ = s.listener.Close()
}

func readClientFrame(t *testing.T, reader *bufio.Reader) string {
	t.Helper()
	header := make([]byte, 2)
	if _, err := io.ReadFull(reader, header); err != nil {
		t.Fatalf("read frame header: %v", err)
	}
	if header[0]&0x0f != 0x1 {
		t.Fatalf("opcode = %d, want text", header[0]&0x0f)
	}
	if header[1]&0x80 == 0 {
		t.Fatal("client frame was not masked")
	}
	length := int(header[1] & 0x7f)
	if length == 126 {
		extended := make([]byte, 2)
		if _, err := io.ReadFull(reader, extended); err != nil {
			t.Fatalf("read extended length: %v", err)
		}
		length = int(binary.BigEndian.Uint16(extended))
	}
	mask := make([]byte, 4)
	if _, err := io.ReadFull(reader, mask); err != nil {
		t.Fatalf("read mask: %v", err)
	}
	payload := make([]byte, length)
	if _, err := io.ReadFull(reader, payload); err != nil {
		t.Fatalf("read payload: %v", err)
	}
	for i := range payload {
		payload[i] ^= mask[i%4]
	}
	return string(payload)
}

func writeServerFrame(t *testing.T, conn net.Conn, opcode byte, payload []byte) {
	t.Helper()
	header := []byte{0x80 | opcode}
	if len(payload) < 126 {
		header = append(header, byte(len(payload)))
	} else {
		header = append(header, 126, 0, 0)
		binary.BigEndian.PutUint16(header[len(header)-2:], uint16(len(payload)))
	}
	if _, err := conn.Write(header); err != nil {
		t.Fatalf("write header: %v", err)
	}
	if _, err := conn.Write(payload); err != nil {
		t.Fatalf("write payload: %v", err)
	}
}

func testAcceptKey(key string) string {
	sum := sha1.Sum([]byte(key + wsGUID))
	return base64.StdEncoding.EncodeToString(sum[:])
}
