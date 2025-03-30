package stream

// Message represents a message that can be sent through a stream.
type Message struct {
	// Key is the message key.
	Key string
	// Metadata contains additional metadata about the message.
	Metadata map[string]string
	// Data is the message data.
	Data []byte
}
