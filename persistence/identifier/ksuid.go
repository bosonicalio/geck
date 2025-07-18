package identifier

import (
	"github.com/segmentio/ksuid"
)

// FactoryKSUID is the concrete implementation of [Factory] using K-Sortable Unique Identifier (KSUID)
// algorithm.
type FactoryKSUID struct{}

// compile-time assertions
var _ Factory = (*FactoryKSUID)(nil)

func (f FactoryKSUID) NewID() (string, error) {
	id, err := ksuid.NewRandom()
	if err != nil {
		return "", err
	}
	return id.String(), nil
}
