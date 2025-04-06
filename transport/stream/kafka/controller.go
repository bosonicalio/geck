package kafka

// A Controller is a component used to register readers to a [ReaderManager].
//
// This way, [ReaderManager] will be able to start the readers and process the messages from Kafka.
type Controller interface {
	// RegisterReaders registers the readers to the [ReaderManager].
	RegisterReaders(manager ReaderManager)
}
