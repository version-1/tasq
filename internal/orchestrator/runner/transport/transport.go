package transport

import "context"

// Connection carries JSON-RPC frames between the runner and an app-server.
type Connection interface {
	Identifier() string
	FrameSource() string
	Send(ctx context.Context, frame []byte) error
	Receive(ctx context.Context) ([]byte, error)
	Close() error
	Done() <-chan error
}
