package event

import "context"

type Writer interface {
	Write(ctx context.Context, events []Event) error
}
