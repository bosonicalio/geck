package kafkatest

import (
	"context"
	"fmt"
	"testing"

	"github.com/samber/lo"
	"github.com/testcontainers/testcontainers-go"
	testcontainerskafka "github.com/testcontainers/testcontainers-go/modules/kafka"
)

// Container represents a Kafka container for testing.
type Container struct {
	Instance        testcontainers.Container
	SeedBrokerAddrs []string
}

// NewContainer creates and starts a Kafka container with configurations for testing scenarios.
func NewContainer(ctx context.Context, t *testing.T, opts ...ContainerOption) (*Container, error) {
	t.Helper()

	options := containerOptions{}
	for _, opt := range opts {
		opt(&options)
	}

	instance, err := testcontainerskafka.Run(ctx,
		fmt.Sprintf("confluentinc/confluent-local:%s", lo.CoalesceOrEmpty(options.imageTag, "7.5.0")),
		testcontainerskafka.WithClusterID("test-cluster"),
	)
	if err != nil {
		return nil, err
	}

	brokerAddrs, err := instance.Brokers(ctx)
	if err != nil {
		return nil, err
	}
	return &Container{
		Instance:        instance,
		SeedBrokerAddrs: brokerAddrs,
	}, nil
}

// -- Option(s) --

type containerOptions struct {
	imageTag string
}

// ContainerOption represents an option for the container.
type ContainerOption func(*containerOptions)

// WithContainerImageTag sets the image tag for the container.
func WithContainerImageTag(imageTag string) ContainerOption {
	return func(o *containerOptions) {
		o.imageTag = imageTag
	}
}
