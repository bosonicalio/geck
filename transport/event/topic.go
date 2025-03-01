package event

import (
	"fmt"
	"strings"
)

type Topic struct {
	Organization string
	Platform     string
	Entity       string
	Action       string
}

// compile-time assertion
var _ fmt.Stringer = (*Topic)(nil)

func NewTopic(entity, action string, opts ...TopicOption) Topic {
	topic := Topic{
		Entity: entity,
		Action: action,
	}
	for _, o := range opts {
		o(&topic)
	}
	return topic
}

func (t Topic) String() string {
	return strings.Join(
		[]string{
			t.Organization,
			t.Platform,
			t.Entity,
			t.Action,
		}, ".")
}

// -- Options --

type TopicOption func(o *Topic)

func WithOrganization(org string) TopicOption {
	return func(o *Topic) {
		o.Organization = org
	}
}

func WithPlatform(platform string) TopicOption {
	return func(o *Topic) {
		o.Platform = platform
	}
}
