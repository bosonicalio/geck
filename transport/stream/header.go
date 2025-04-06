package stream

import "net/textproto"

// Header is a map of string key-value pairs representing message headers.
type Header map[string][]string

// Set sets the value associated with the given key in the message header.
func (h Header) Set(key, value string) {
	textproto.MIMEHeader(h).Set(key, value)
}

// Add adds a value to the list of values associated with the given key in the message header.
func (h Header) Add(key, value string) {
	textproto.MIMEHeader(h).Add(key, value)
}

// Get returns the value associated with the given key in the message header.
func (h Header) Get(key string) string {
	return textproto.MIMEHeader(h).Get(key)
}

// Values returns all values associated with the given key in the message header.
func (h Header) Values(key string) []string {
	return textproto.MIMEHeader(h).Values(key)
}
