package awstest

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/localstack"

	"github.com/tesserical/geck/testutil"
)

// Pod is a test component for running a LocalStack docker instance with required Amazon Web Services (AWS) services.
type Pod struct {
	baseCtx   context.Context
	container *localstack.LocalStackContainer
	config    aws.Config
}

// compile-time assertions
var _ testutil.Pod = (*Pod)(nil)

// NewPod creates a new LocalStack container configured to run any given services.
//
// It accepts options to specify the Docker image tag and the services to be started.
//
// The default image tag is "latest", and if no services are specified, it will run with all services.
func NewPod(ctx context.Context, opts ...PodOption) (Pod, error) {
	podConfig := newPodOptions()
	for _, opt := range opts {
		opt(podConfig)
	}
	envMap := map[string]string{}
	if len(podConfig.services) > 0 {
		envMap["SERVICES"] = strings.Join(podConfig.services, ",")
	}
	container, err := localstack.Run(ctx, fmt.Sprintf("localstack/localstack:%s", podConfig.imageTag),
		testcontainers.WithEnv(envMap),
	)
	if err != nil {
		return Pod{}, err
	}
	defer func() {
		// Ensure the container is terminated if an error occurs
		if err != nil && container != nil {
			_ = container.Terminate(ctx)
		}
	}()
	mappedPort, err := container.MappedPort(ctx, "4566/tcp")
	if err != nil {
		return Pod{}, err
	}
	provider, err := testcontainers.NewDockerProvider()
	if err != nil {
		return Pod{}, err
	}
	defer func() {
		_ = provider.Close() // Ensure the provider is closed after use
	}()
	host, err := provider.DaemonHost(ctx)
	if err != nil {
		return Pod{}, err
	}
	baseEndpoint := fmt.Sprintf("http://%s:%s", host, mappedPort.Port())
	// Create the AWS configuration with the LocalStack endpoint
	awsConfig, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("some_key", "some_secret", "")),
		config.WithBaseEndpoint(baseEndpoint),
	)
	return Pod{
		baseCtx:   ctx,
		container: container,
		config:    awsConfig,
	}, nil
}

// Close terminates the LocalStack container.
func (p Pod) Close() error {
	if p.container == nil {
		return nil // No container to terminate
	}
	return p.container.Terminate(p.baseCtx)
}

// Config returns the AWS configuration for the LocalStack container.
func (p Pod) Config() aws.Config {
	return p.config
}

// -- Options --

type podOptions struct {
	imageTag string
	services []string
}

func newPodOptions() *podOptions {
	return &podOptions{
		imageTag: "latest",   // Default image tag
		services: []string{}, // Default to no specific services
	}
}

// PodOption is a functional option type for configuring the LocalStack Pod.
type PodOption func(*podOptions)

// WithPodImageTag sets the Docker image tag for the LocalStack container.
func WithPodImageTag(tag string) PodOption {
	return func(opts *podOptions) {
		opts.imageTag = tag
	}
}

// WithPodServices sets the services to be started in the LocalStack container.
func WithPodServices(services ...string) PodOption {
	return func(opts *podOptions) {
		opts.services = services
	}
}
