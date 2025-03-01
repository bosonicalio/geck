package event

type Aggregator interface {
	RegisterEvents(events ...Event)
	PullEvents() []Event
}

type AggregatorTemplate struct {
	events []Event
}

// compile-time assertion
var _ Aggregator = (*AggregatorTemplate)(nil)

func (t *AggregatorTemplate) RegisterEvents(events ...Event) {
	if t.events == nil {
		t.events = make([]Event, 0, len(events))
	}
	t.events = append(t.events, events...)
}

func (t AggregatorTemplate) PullEvents() []Event {
	return t.events
}
