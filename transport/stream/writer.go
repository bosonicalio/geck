package stream

import "context"

// Writer is a component that writes messages to a stream.
type Writer interface {
	// Write writes a message to the stream.
	Write(ctx context.Context, name string, message Message) error
	// WriteBatch writes a batch of messages to the stream.
	WriteBatch(ctx context.Context, name string, messages []Message) (int, error)
}
