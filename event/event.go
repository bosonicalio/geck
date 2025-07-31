package event

import (
	"time"

	"github.com/bosonicalio/geck/transport"
)

// CloudEventsCurrentSpecVersion is the current version of the Cloud Events specification.
const CloudEventsCurrentSpecVersion = "1.0"

// An Event is a fact that something happened in the system.
//
// Events are mostly streamed to the interested parties (subscribers) to notify them about the changes without
// making them dependent to the broadcasting system (fire-and-forget).
//
// This interface is based on the Cloud Events specification.
//
// More information about the Cloud Events specification can be found at:
// https://github.com/cloudevents/spec.
type Event interface {
	// Topic returns the [Topic] of the event.
	Topic() Topic
	// Key returns the key of the event.
	//
	// Depending on the implementation, the key can be used to route the given event
	// into a certain data sub-set (partition/shard). So, be careful when defining it.
	//
	// If no key is needed, just return an empty string.
	Key() string
	// Bytes returns the serialized representation of the event.
	Bytes() ([]byte, error)
	// BytesContentType returns the content MIME type ([transport.MimeType]) of the serialized bytes.
	BytesContentType() transport.MimeType
	// Source identifies the context in which a message was produced.
	Source() string
	// Subject describes the subject of the event in the context of the event producer (identified by Source).
	Subject() string
	// OccurrenceTime returns the occurrence time of the event.
	OccurrenceTime() time.Time
	// SchemaSource returns the schema source of the event data.
	SchemaSource() string
}
