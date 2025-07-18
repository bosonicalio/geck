package identifier

import "github.com/google/uuid"

// FactoryUUID is a factory for generating UUIDs, concrete implementation of [Factory].
type FactoryUUID struct{}

// compile-time assertion
var _ Factory = (*FactoryUUID)(nil)

func (f FactoryUUID) NewID() (string, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return "", err
	}
	return id.String(), nil
}
