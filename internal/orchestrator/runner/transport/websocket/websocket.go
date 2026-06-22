package websocket

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/sha1"
	"crypto/tls"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	maxFrameSize = 10 * 1024 * 1024
	wsGUID       = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
)

type Connection struct {
	conn   net.Conn
	reader *bufio.Reader
	url    string
	mu     sync.Mutex
	closed bool
}

func Dial(ctx context.Context, rawURL string) (*Connection, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("parse websocket url: %w", err)
	}
	if parsed.Scheme != "ws" && parsed.Scheme != "wss" {
		return nil, fmt.Errorf("unsupported websocket scheme %q", parsed.Scheme)
	}
	host := parsed.Host
	if !strings.Contains(host, ":") {
		if parsed.Scheme == "wss" {
			host += ":443"
		} else {
			host += ":80"
		}
	}
	dialer := net.Dialer{}
	var conn net.Conn
	if parsed.Scheme == "wss" {
		tcpConn, dialErr := dialer.DialContext(ctx, "tcp", host)
		if dialErr != nil {
			return nil, fmt.Errorf("connect websocket: %w", dialErr)
		}
		tlsConn := tls.Client(tcpConn, &tls.Config{ServerName: parsed.Hostname()})
		if err := tlsConn.HandshakeContext(ctx); err != nil {
			_ = tcpConn.Close()
			return nil, fmt.Errorf("connect websocket tls: %w", err)
		}
		conn = tlsConn
	} else {
		conn, err = dialer.DialContext(ctx, "tcp", host)
	}
	if err != nil {
		return nil, fmt.Errorf("connect websocket: %w", err)
	}
	client := &Connection{conn: conn, reader: bufio.NewReader(conn), url: rawURL}
	if err := client.handshake(ctx, parsed); err != nil {
		_ = conn.Close()
		return nil, err
	}
	return client, nil
}

func (c *Connection) Identifier() string {
	return "websocket=" + c.url
}

func (c *Connection) FrameSource() string {
	return "websocket"
}

func (c *Connection) Send(ctx context.Context, frame []byte) error {
	return c.writeFrame(ctx, 0x1, frame)
}

func (c *Connection) Receive(ctx context.Context) ([]byte, error) {
	var message []byte
	for {
		opcode, fin, payload, err := c.readFrame(ctx)
		if err != nil {
			return nil, err
		}
		switch opcode {
		case 0x0, 0x1:
			message = append(message, payload...)
			if len(message) > maxFrameSize {
				return nil, fmt.Errorf("websocket frame exceeds %d bytes", maxFrameSize)
			}
			if fin {
				return message, nil
			}
		case 0x8:
			return nil, io.EOF
		case 0x9:
			if err := c.writeFrame(ctx, 0xA, payload); err != nil {
				return nil, err
			}
		case 0xA:
			continue
		default:
			return nil, fmt.Errorf("unsupported websocket opcode %d", opcode)
		}
	}
}

func (c *Connection) Close() error {
	c.mu.Lock()
	alreadyClosed := c.closed
	c.closed = true
	c.mu.Unlock()
	if !alreadyClosed {
		_ = c.writeFrame(context.Background(), 0x8, nil)
	}
	return c.conn.Close()
}

func (c *Connection) Done() <-chan error {
	return nil
}

func (c *Connection) handshake(ctx context.Context, parsed *url.URL) error {
	key, err := nonce()
	if err != nil {
		return err
	}
	path := parsed.RequestURI()
	if path == "" {
		path = "/"
	}
	if err := setConnDeadline(ctx, c.conn); err != nil {
		return err
	}
	defer clearDeadline(c.conn)
	req := fmt.Sprintf(
		"GET %s HTTP/1.1\r\nHost: %s\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Key: %s\r\nSec-WebSocket-Version: 13\r\n\r\n",
		path,
		parsed.Host,
		key,
	)
	if _, err := io.WriteString(c.conn, req); err != nil {
		return fmt.Errorf("send websocket handshake: %w", err)
	}
	resp, err := http.ReadResponse(c.reader, &http.Request{Method: http.MethodGet})
	if err != nil {
		return fmt.Errorf("read websocket handshake: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusSwitchingProtocols {
		return fmt.Errorf("websocket handshake status %s", resp.Status)
	}
	if !headerToken(resp.Header.Get("Upgrade"), "websocket") || !headerToken(resp.Header.Get("Connection"), "upgrade") {
		return errors.New("websocket handshake missing upgrade headers")
	}
	if got, want := resp.Header.Get("Sec-WebSocket-Accept"), acceptKey(key); got != want {
		return errors.New("websocket handshake accept mismatch")
	}
	return nil
}

func (c *Connection) writeFrame(ctx context.Context, opcode byte, payload []byte) error {
	if len(payload) > maxFrameSize {
		return fmt.Errorf("websocket frame exceeds %d bytes", maxFrameSize)
	}
	if err := setDeadline(ctx, c.conn, true); err != nil {
		return err
	}
	defer clearDeadline(c.conn)
	header := []byte{0x80 | opcode}
	switch {
	case len(payload) < 126:
		header = append(header, 0x80|byte(len(payload)))
	case len(payload) <= 0xffff:
		header = append(header, 0x80|126, 0, 0)
		binary.BigEndian.PutUint16(header[len(header)-2:], uint16(len(payload)))
	default:
		header = append(header, 0x80|127, 0, 0, 0, 0, 0, 0, 0, 0)
		binary.BigEndian.PutUint64(header[len(header)-8:], uint64(len(payload)))
	}
	mask := make([]byte, 4)
	if _, err := rand.Read(mask); err != nil {
		return fmt.Errorf("generate websocket mask: %w", err)
	}
	masked := make([]byte, len(payload))
	for i := range payload {
		masked[i] = payload[i] ^ mask[i%4]
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, err := c.conn.Write(header); err != nil {
		return fmt.Errorf("write websocket frame header: %w", err)
	}
	if _, err := c.conn.Write(mask); err != nil {
		return fmt.Errorf("write websocket frame mask: %w", err)
	}
	if _, err := c.conn.Write(masked); err != nil {
		return fmt.Errorf("write websocket frame payload: %w", err)
	}
	return nil
}

func (c *Connection) readFrame(ctx context.Context) (byte, bool, []byte, error) {
	if err := setDeadline(ctx, c.conn, false); err != nil {
		return 0, false, nil, err
	}
	defer clearDeadline(c.conn)
	header := make([]byte, 2)
	if _, err := io.ReadFull(c.reader, header); err != nil {
		return 0, false, nil, normalizeReadError(err)
	}
	fin := header[0]&0x80 != 0
	opcode := header[0] & 0x0f
	masked := header[1]&0x80 != 0
	length := uint64(header[1] & 0x7f)
	switch length {
	case 126:
		extended := make([]byte, 2)
		if _, err := io.ReadFull(c.reader, extended); err != nil {
			return 0, false, nil, normalizeReadError(err)
		}
		length = uint64(binary.BigEndian.Uint16(extended))
	case 127:
		extended := make([]byte, 8)
		if _, err := io.ReadFull(c.reader, extended); err != nil {
			return 0, false, nil, normalizeReadError(err)
		}
		length = binary.BigEndian.Uint64(extended)
	}
	if length > maxFrameSize {
		return 0, false, nil, fmt.Errorf("websocket frame exceeds %d bytes", maxFrameSize)
	}
	var mask []byte
	if masked {
		mask = make([]byte, 4)
		if _, err := io.ReadFull(c.reader, mask); err != nil {
			return 0, false, nil, normalizeReadError(err)
		}
	}
	payload := make([]byte, length)
	if _, err := io.ReadFull(c.reader, payload); err != nil {
		return 0, false, nil, normalizeReadError(err)
	}
	for i := range payload {
		if masked {
			payload[i] ^= mask[i%4]
		}
	}
	return opcode, fin, payload, nil
}

func nonce() (string, error) {
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generate websocket key: %w", err)
	}
	return base64.StdEncoding.EncodeToString(raw), nil
}

func acceptKey(key string) string {
	sum := sha1.Sum([]byte(key + wsGUID))
	return base64.StdEncoding.EncodeToString(sum[:])
}

func headerToken(value string, token string) bool {
	for _, part := range strings.Split(value, ",") {
		if strings.EqualFold(strings.TrimSpace(part), token) {
			return true
		}
	}
	return false
}

func setDeadline(ctx context.Context, conn net.Conn, write bool) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	deadline, ok := ctx.Deadline()
	if !ok {
		return nil
	}
	if write {
		return conn.SetWriteDeadline(deadline)
	}
	return conn.SetReadDeadline(deadline)
}

func setConnDeadline(ctx context.Context, conn net.Conn) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	deadline, ok := ctx.Deadline()
	if !ok {
		return nil
	}
	return conn.SetDeadline(deadline)
}

func clearDeadline(conn net.Conn) {
	_ = conn.SetDeadline(time.Time{})
}

func normalizeReadError(err error) error {
	if errors.Is(err, io.EOF) {
		return io.EOF
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return context.DeadlineExceeded
	}
	return err
}
