package transport

// Message is a sentinel structure to hold a sole item as Data.
//
// This structure helps APIs exposed to external system clients through transport protocols
// to create a definition of a returning values and thus, a clear separation of different return values
// (i.e. errors).
type Message[T any] struct {
	// Data the item to be tagged as data value.
	Data T `json:"data"`
}
