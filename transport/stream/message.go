package stream

// Message represents a message that can be sent through a stream.
type Message struct {
	// Key is the message key.
	Key string
	// Header contains additional metadata about the message.
	Header Header
	// Data is the message data.
	Data []byte
}
