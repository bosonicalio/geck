package kafka

import (
	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/tesserical/geck/transport/stream"
)

func marshalHeaders(headers stream.Header) []kgo.RecordHeader {
	kgoHeaders := make([]kgo.RecordHeader, 0, len(headers))
	for k := range headers {
		kgoHeaders = append(kgoHeaders, kgo.RecordHeader{
			Key:   k,
			Value: []byte(headers.Get(k)),
		})
	}
	return kgoHeaders
}

// ParseHeaders parses the headers from a Kafka record into a map.
func ParseHeaders(record *kgo.Record) stream.Header {
	m := make(stream.Header, len(record.Headers))
	for _, h := range record.Headers {
		m.Add(h.Key, string(h.Value))
	}
	return m
}
