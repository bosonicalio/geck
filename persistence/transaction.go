package persistence

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// Transaction is a database transaction interface that provides methods for
// managing transaction lifecycle. Implementations should handle the underlying
// database-specific transaction operations while maintaining consistent behavior.
// Use with [WithTxContext] and [FromTxContext] to propagate transactions through context.
type Transaction interface {
	// Commit commits the transaction.
	Commit(ctx context.Context) error
	// Rollback rolls back the transaction.
	Rollback(ctx context.Context) error
}

// -- Context Management --

// TxExecutor is a type that represents the executor of a transaction.
//
// It is used as a key in the context to store and retrieve the [Transaction] instance.
type TxExecutor string

var (
	// ErrInvalidTxContext is returned when a transaction context is invalid or does not contain a transaction.
	ErrInvalidTxContext = errors.New("geck.persistence: invalid transaction context")
)

// WithTxContext creates a new context with the given [Transaction].
func WithTxContext(parent context.Context, executor TxExecutor, tx Transaction) context.Context {
	return context.WithValue(parent, executor, tx)
}

// FromTxContext retrieves the [Transaction] from the given context.
func FromTxContext(ctx context.Context, executor TxExecutor) (Transaction, bool) {
	tx, ok := ctx.Value(executor).(Transaction)
	return tx, ok
}

// ExecInTx executes the provided function `fn` within the context of a transaction.
// It automatically handles transaction commit and rollback based on the function's execution result.
// If `fn` panics, it will recover and rollback the transaction, returning the panic as an error.
// The function `fn` receives a context that has the transaction set, allowing it to perform database operations
// within the transaction scope.
func ExecInTx(ctx context.Context, factory TxFactory, fn func(ctx context.Context) error) (err error) {
	tx, err := factory.NewTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to create new transaction: %w", err)
	}
	txCtx := WithTxContext(ctx, factory.Executor(), tx)
	defer func() {
		if r := recover(); r != nil {
			// Ensure we handle panics gracefully starting from persistence layer
			panicErr, ok := r.(error)
			if !ok {
				panicErr = fmt.Errorf("transaction panic: %v", r)
			}
			err = errors.Join(err, panicErr)
		}
		if err != nil {
			if errRollback := tx.Rollback(txCtx); errRollback != nil {
				err = errors.Join(err, errRollback)
			}
			return
		}
		if errCommit := tx.Commit(txCtx); errCommit != nil {
			err = errors.Join(err, errCommit)
		}
	}()
	err = fn(txCtx)
	return
}

// ExecInTxAll executes the provided function `fn` within the context of transactions from all registered factories.
// It creates transactions from all registered factories and coordinates them as a single logical transaction.
//
// If any transaction fails during commit, all transactions are rolled back to maintain consistency.
// The function `fn` receives a context that has all transactions set, allowing it to perform database operations
// across multiple transaction contexts.
func ExecInTxAll(ctx context.Context, fn func(ctx context.Context) error) (err error) {
	factories := GetTxFactories()
	if len(factories) == 0 {
		return errors.New("no transaction factories registered")
	}

	type txInfo struct {
		tx       Transaction
		executor TxExecutor
	}

	transactions := make([]txInfo, 0, len(factories))
	txCtx := ctx

	// Create transactions from all factories
	for _, factory := range factories {
		tx, err := factory.NewTx(ctx)
		if err != nil {
			// Rollback any already created transactions
			for _, txInfo := range transactions {
				_ = txInfo.tx.Rollback(ctx)
			}
			return fmt.Errorf("failed to create transaction for executor %s: %w", factory.Executor(), err)
		}
		transactions = append(transactions, txInfo{tx: tx, executor: factory.Executor()})
		txCtx = WithTxContext(txCtx, factory.Executor(), tx)
	}

	defer func() {
		if r := recover(); r != nil {
			panicErr, ok := r.(error)
			if !ok {
				panicErr = fmt.Errorf("transaction panic: %v", r)
			}
			err = errors.Join(err, panicErr)
		}

		if err != nil {
			// Rollback all transactions in reverse order
			for i := len(transactions) - 1; i >= 0; i-- {
				if errRollback := transactions[i].tx.Rollback(txCtx); errRollback != nil {
					err = errors.Join(err, errRollback)
				}
			}
			return
		}

		// Commit all transactions - if any fails, rollback all
		for _, txInfo := range transactions {
			if errCommit := txInfo.tx.Commit(txCtx); errCommit != nil {
				err = errors.Join(err, errCommit)
				// Rollback remaining transactions
				for j := len(transactions) - 1; j >= 0; j-- {
					if errRollback := transactions[j].tx.Rollback(txCtx); errRollback != nil {
						err = errors.Join(err, errRollback)
					}
				}
				return
			}
		}
	}()

	err = fn(txCtx)
	return
}

// --- Factory ---

var (
	_txFactoryMu       = &sync.Mutex{}
	_txFactoryRegistry []TxFactory
)

// TxFactory is an interface that defines methods for creating and managing transactions.
//
// This factories must be registered using [RegisterTxFactory] before they can be used inside the application.
// This way, multiple transaction factories can be registered, allowing the application to support different database
// backends or configurations.
//
// The Persistence API will hold a global registry of all registered transaction factories, which can be accessed using
// [GetTxFactories]. Each factory can create its own [Transaction] instances, which can be used to execute database operations
// within a transaction scope.
type TxFactory interface {
	// Executor returns the [TxExecutor] associated with this factory.
	Executor() TxExecutor
	// NewTx creates a new [Transaction] instance.
	NewTx(ctx context.Context) (Transaction, error)
}

// RegisterTxFactory registers a new [TxFactory] instance into the global registry.
func RegisterTxFactory(factory TxFactory) {
	_txFactoryMu.Lock()
	defer _txFactoryMu.Unlock()
	if _txFactoryRegistry == nil {
		_txFactoryRegistry = make([]TxFactory, 0, 1)
	}
	_txFactoryRegistry = append(_txFactoryRegistry, factory)
}

// GetTxFactories returns a slice of all [TxFactory] instances registered in the global registry.
func GetTxFactories() []TxFactory {
	_txFactoryMu.Lock()
	defer _txFactoryMu.Unlock()
	if _txFactoryRegistry == nil {
		return nil
	}
	factories := make([]TxFactory, len(_txFactoryRegistry))
	copy(factories, _txFactoryRegistry)
	return factories
}
