package kafka

import (
	"errors"
	"fmt"
	"strings"
)

const (
	_groupNameFormat          = "%s.%s.%s"
	_groupNameWithEventFormat = "%s.%s.%s-on-%s"
)

// A ConsumerGroup is an Apache Kafka logical grouping of consumers that commit offsets in one or several topics
// in a coordinated manner. It is used to ensure that each message is processed only once by a single consumer in
// the group, perfectly for clustered/multi-node applications.
type ConsumerGroup struct {
	platform string
	service  string
	task     string
	event    string
}

// compile-time assertion
var _ fmt.Stringer = (*ConsumerGroup)(nil)

// NewConsumerGroup creates a new instance of [ConsumerGroup].
//
// All fields not marked as optional are required.
func NewConsumerGroup(platform, service, task string, opts ...ConsumerGroupOption) (ConsumerGroup, error) {
	group := ConsumerGroup{
		platform: platform,
		service:  service,
		task:     task,
	}
	if group.IsZero() {
		return ConsumerGroup{}, errors.New("consumer group is missing a required field")
	}
	for _, opt := range opts {
		opt(&group)
	}
	return group, nil
}

// MustConsumerGroup creates a new instance of [ConsumerGroup].
//
// All fields not marked as optional are required. If a required field is not set, this routine will panic.
func MustConsumerGroup(platform, service, task string, opts ...ConsumerGroupOption) ConsumerGroup {
	group, err := NewConsumerGroup(platform, service, task, opts...)
	if err != nil {
		panic(err)
	}
	return group
}

// ParseConsumerGroup parses a consumer group name and returns a [ConsumerGroup].
func ParseConsumerGroup(groupName string) (ConsumerGroup, error) {
	parts := strings.Split(groupName, ".")
	if len(parts) < 3 {
		return ConsumerGroup{}, errors.New("invalid consumer group name")
	}

	group := ConsumerGroup{
		platform: parts[0],
		service:  parts[1],
		task:     parts[2],
	}
	if len(parts) == 4 {
		group.event = parts[3]
	}
	return group, nil
}

// MustParseConsumerGroup parses a consumer group name and returns a [ConsumerGroup].
//
// This routine will panic if the group name is invalid.
func MustParseConsumerGroup(groupName string) ConsumerGroup {
	group, err := ParseConsumerGroup(groupName)
	if err != nil {
		panic(err)
	}
	return group
}

// IsZero checks if the consumer group is empty.
func (c ConsumerGroup) IsZero() bool {
	return c.platform == "" || c.service == "" || c.task == ""
}

// Platform returns the platform name of the consumer group.
func (c ConsumerGroup) Platform() string {
	return c.platform
}

// Service returns the service name of the consumer group.
func (c ConsumerGroup) Service() string {
	return c.service
}

// Task returns the task name of the consumer group.
func (c ConsumerGroup) Task() string {
	return c.task
}

// Event returns the event name of the consumer group.
func (c ConsumerGroup) Event() string {
	return c.event
}

// GroupName returns the group name of the consumer group.
//
// The name convention is: [platform-name].[service-name].[task-name].
// If the event name is set, the format will be: [platform-name].[service-name].[task-name]-on-[event-name].
func (c ConsumerGroup) String() string {
	if c.IsZero() {
		return ""
	} else if c.event != "" {
		return fmt.Sprintf(_groupNameWithEventFormat, c.platform, c.service, c.task, c.event)
	}
	return fmt.Sprintf(_groupNameFormat, c.platform, c.service, c.task)
}

// -- Options --

// ConsumerGroupOption is a function that modifies the consumer group.
type ConsumerGroupOption func(*ConsumerGroup)

// WithConsumerGroupEvent sets the event name for the consumer group.
//
// This represents the event a consumer group is listening to and thus, performing the task.
func WithConsumerGroupEvent(event string) ConsumerGroupOption {
	return func(cg *ConsumerGroup) {
		cg.event = event
	}
}
