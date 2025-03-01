package sql

import (
	"context"
	"database/sql"

	"github.com/hadroncorp/geck/persistence"
)

// Transaction is an adapter structure for [persistence.Transaction].
type Transaction struct {
	Parent *sql.Tx
}

// compile-time assertion
var _ persistence.Transaction = (*Transaction)(nil)

func (t Transaction) Commit(_ context.Context) error {
	return t.Parent.Commit()
}

func (t Transaction) Rollback(_ context.Context) error {
	return t.Parent.Rollback()
}
