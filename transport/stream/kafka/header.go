package kafka

import (
	"strconv"

	"github.com/twmb/franz-go/pkg/kgo"
)

const (
	// HeaderKeyKafkaTopic is the key for the Kafka topic header.
	HeaderKeyKafkaTopic = "kafka-topic"
	// HeaderKeyKafkaPartition is the key for the Kafka partition header.
	HeaderKeyKafkaPartition = "kafka-partition"
	// HeaderKeyKafkaOffset is the key for the Kafka offset header.
	HeaderKeyKafkaOffset = "kafka-offset"
	// HeaderKeyKafkaTimestamp is the key for the Kafka timestamp header.
	HeaderKeyKafkaTimestamp = "kafka-timestamp"
	// HeaderKeyKafkaTimestampType is the key for the Kafka timestamp type header.
	HeaderKeyKafkaTimestampType = "kafka-timestamp-type"
	// HeaderKeyCompressionType is the key for the compression type header.
	HeaderKeyCompressionType = "compression-type"
)

func marshalHeaders(headers map[string]string) []kgo.RecordHeader {
	kgoHeaders := make([]kgo.RecordHeader, 0, len(headers))
	for k, v := range headers {
		kgoHeaders = append(kgoHeaders, kgo.RecordHeader{
			Key:   k,
			Value: []byte(v),
		})
	}
	return kgoHeaders
}

func unmarshalHeaders(record *kgo.Record) map[string]string {
	const totalKafkaHeaders = 7
	m := make(map[string]string, len(record.Headers)+totalKafkaHeaders)
	for _, h := range record.Headers {
		m[h.Key] = string(h.Value)
	}
	m[HeaderKeyKafkaTopic] = record.Topic
	m[HeaderKeyKafkaPartition] = strconv.Itoa(int(record.Partition))
	m[HeaderKeyKafkaOffset] = strconv.Itoa(int(record.Offset))
	m[HeaderKeyKafkaTimestamp] = record.Timestamp.String()
	m[HeaderKeyKafkaTimestampType] = strconv.Itoa(int(record.Attrs.TimestampType()))
	m[HeaderKeyCompressionType] = strconv.Itoa(int(record.Attrs.CompressionType()))
	return m
}
