package sql

import (
	"context"
	"database/sql"

	"github.com/tesserical/geck/persistence"
)

// TxDriver is a type alias for the [persistence.TxDriver] used in the context of a [DBTxPropagator].
const TxDriver persistence.TxDriver = "sql"

// Transaction is the adapter structure of [persistence.Transaction] for [sql].
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

// -- Factory --

// TxFactory is the concrete implementation of [persistence.TxFactory] for [sql].
type TxFactory struct {
	client DB
	opts   *sql.TxOptions
}

// NewTxFactory creates a new instance of [TxFactory] with the provided [DB] client and transaction options.
func NewTxFactory(client DB, txOpts *sql.TxOptions) TxFactory {
	return TxFactory{
		client: client,
		opts:   txOpts,
	}
}

// compile-time assertion
var _ persistence.TxFactory = (*TxFactory)(nil)

func (t TxFactory) Driver() persistence.TxDriver {
	return TxDriver
}

func (t TxFactory) NewTx(ctx context.Context) (persistence.Transaction, error) {
	tx, err := t.client.BeginTx(ctx, t.opts)
	if err != nil {
		return nil, err
	}
	return Transaction{Parent: tx}, nil
}
