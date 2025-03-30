package event

import (
	"context"
	"time"

	"github.com/hadroncorp/geck/persistence/identifier"
	"github.com/hadroncorp/geck/transport/stream"
)

// A Publisher is a component that propagates events to potentially one or more external systems.
//
// This component is different from the one provided by the [stream] package, as it is used to propagate
// events to external systems, not just messages (higher level).
type Publisher interface {
	// Publish propagates the given events.
	Publish(ctx context.Context, events []Event) error
}

// StreamPublisher is a [Publisher] implementation that propagates events to a stream.
//
// This component publishes events into a stream in a synchronous-way using stream write batch APIs.
type StreamPublisher struct {
	writer    stream.Writer
	idFactory identifier.Factory
}

// compile-time assertion(s)
var _ Publisher = (*StreamPublisher)(nil)

// NewStreamPublisher creates a new [StreamPublisher] instance.
func NewStreamPublisher(w stream.Writer, factory identifier.Factory) StreamPublisher {
	return StreamPublisher{writer: w, idFactory: factory}
}

// Publish propagates the given events.
func (p StreamPublisher) Publish(ctx context.Context, events []Event) error {
	topicMessages := make(map[string][]stream.Message)
	for _, event := range events {
		id, err := p.idFactory.NewID()
		if err != nil {
			return err
		}
		topic := event.Topic().String()
		msg, err := event.Bytes()
		if err != nil {
			return err
		}
		topicMessages[topic] = append(topicMessages[topic], stream.Message{
			Key:  event.Key(),
			Data: msg,
			Metadata: map[string]string{
				HeaderEventID:         id,
				HeaderSource:          event.Source(),
				HeaderSpecVersion:     CloudEventsCurrentSpecVersion,
				HeaderEventType:       event.Topic().String(),
				HeaderDataContentType: event.BytesContentType().String(),
				HeaderDataSchema:      event.SchemaSource(),
				HeaderSubject:         event.Subject(),
				HeaderTime:            event.OccurrenceTime().Format(time.RFC3339),
			},
		})
	}

	for topic, messages := range topicMessages {
		if _, err := p.writer.WriteBatch(ctx, topic, messages); err != nil {
			return err
		}
	}
	return nil
}
