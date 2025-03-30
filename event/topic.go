package event

import (
	"errors"
	"fmt"
	"strings"

	"github.com/samber/lo"
)

// -- Error(s) --

var (
	// ErrInvalidTopicName is returned when the given topic name is invalid.
	ErrInvalidTopicName = errors.New("invalid topic name")
)

// A Topic is the name of the logical group (aka. channel) a certain set of messages will be broadcast/streamed.
type Topic struct {
	organization string
	platform     string
	entity       string
	action       string

	strVal string
}

// compile-time assertion
var _ fmt.Stringer = (*Topic)(nil)

// NewTopic creates a new Topic with the given entity and action.
//
// It also accepts optionals [TopicOption] to set the organization and platform.
func NewTopic(org, entity, action string, opts ...TopicOption) Topic {
	topic := Topic{
		organization: org,
		entity:       entity,
		action:       action,
	}
	for _, o := range opts {
		o(&topic)
	}
	topic.strVal = strings.Join(
		lo.Compact([]string{
			topic.organization,
			topic.platform,
			topic.entity,
			topic.action,
		}),
		".",
	)
	return topic
}

// ParseTopic parses the given string into a Topic.
func ParseTopic(v string) (Topic, error) {
	topic := Topic{
		strVal: v,
	}

	splitStr := strings.Split(v, ".")
	if len(splitStr) < 3 {
		return topic, ErrInvalidTopicName
	}

	if len(splitStr) == 4 {
		topic.organization = splitStr[0]
		topic.platform = splitStr[1]
		topic.entity = splitStr[2]
		topic.action = splitStr[3]
	} else {
		topic.organization = splitStr[0]
		topic.entity = splitStr[1]
		topic.action = splitStr[2]
	}
	return topic, nil
}

// String returns the string representation of the Topic.
func (t Topic) String() string {
	return t.strVal
}

// Organization returns the organization of the Topic.
func (t Topic) Organization() string {
	return t.organization
}

// Platform returns the platform of the Topic.
func (t Topic) Platform() string {
	return t.platform
}

// Entity returns the entity of the Topic.
func (t Topic) Entity() string {
	return t.entity
}

// Action returns the action of the Topic.
func (t Topic) Action() string {
	return t.action
}

// -- Options --

// TopicOption is a functional option to set the properties of a Topic.
type TopicOption func(o *Topic)

// WithPlatform sets the platform of the Topic.
func WithPlatform(platform string) TopicOption {
	return func(o *Topic) {
		o.platform = platform
	}
}
