package identifier

import "github.com/google/uuid"

// UUIDFactory is a factory for generating UUIDs, concrete implementation of [Factory].
type UUIDFactory struct{}

// compile-time assertion
var _ Factory = (*UUIDFactory)(nil)

// NewUUIDFactory creates a new instance of [UUIDFactory].
func NewUUIDFactory() UUIDFactory {
	return UUIDFactory{}
}

func (f UUIDFactory) NewID() (string, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return "", err
	}
	return id.String(), nil
}
