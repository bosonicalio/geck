package event

// An Aggregator is a type that can register and pull [Event] sets as desired.
//
// Implement this interface to allow an entity to manage event aggregation operations.
type Aggregator interface {
	// RegisterEvents registers the given events into the aggregator.
	RegisterEvents(events ...Event)
	// PullEvents returns the events registered in the aggregator.
	PullEvents() []Event
}

// An AggregatorTemplate is a basic implementation of the [Aggregator] interface.
type AggregatorTemplate struct {
	events []Event
}

// compile-time assertion
var _ Aggregator = (*AggregatorTemplate)(nil)

// RegisterEvents registers the given events into the aggregator.
func (t *AggregatorTemplate) RegisterEvents(events ...Event) {
	if t.events == nil {
		t.events = make([]Event, 0, len(events))
	}
	t.events = append(t.events, events...)
}

// PullEvents returns the events registered in the aggregator.
func (t AggregatorTemplate) PullEvents() []Event {
	return t.events
}
