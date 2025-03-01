package event

type Event interface {
	Topic() Topic
	Key() string
	Bytes() ([]byte, error)
}
