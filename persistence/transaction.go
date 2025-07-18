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

// TxDriver is a type that represents the driver of a transaction.
//
// It is used as a key in the context to store and retrieve the [Transaction] instance.
type TxDriver string

var (
	// ErrInvalidTxContext is returned when a transaction context is invalid or does not contain a transaction.
	ErrInvalidTxContext = errors.New("geck.persistence: invalid transaction context")
)

// WithTxContext creates a new context with the given [Transaction].
func WithTxContext(parent context.Context, driver TxDriver, tx Transaction) context.Context {
	return context.WithValue(parent, driver, tx)
}

// FromTxContext retrieves the [Transaction] from the given context.
func FromTxContext(ctx context.Context, driver TxDriver) (Transaction, bool) {
	tx, ok := ctx.Value(driver).(Transaction)
	return tx, ok
}

// --- Factory ---

// TxFactory is a component responsible for the creation of persistence transactions.
type TxFactory interface {
	// Driver returns the [TxDriver] associated with this factory.
	Driver() TxDriver
	// NewTx creates a new [Transaction] instance.
	NewTx(ctx context.Context) (Transaction, error)
}

// TxManager is a manager for transaction factories that allows registering multiple
// transaction factories and executing functions within the context of all registered transactions.
// It provides a way to coordinate multiple transactions as a single logical transaction.
//
// It is useful for scenarios where multiple databases or transaction sources need to be coordinated
// together, ensuring that all operations across these transactions are either committed or rolled back
// as a single unit of work.
type TxManager struct {
	regMu     sync.RWMutex
	factories []TxFactory
}

// NewTxManager creates a new instance of [TxManager].
func NewTxManager() *TxManager {
	return &TxManager{
		factories: make([]TxFactory, 0),
	}
}

// Register registers a new transaction factory with the manager.
func (m *TxManager) Register(factory TxFactory) {
	m.regMu.Lock()
	defer m.regMu.Unlock()
	m.factories = append(m.factories, factory)
}

// GetFactories returns a slice of all registered transaction factories.
func (m *TxManager) GetFactories() []TxFactory {
	m.regMu.RLock()
	defer m.regMu.RUnlock()
	if m.factories == nil {
		return nil
	}
	return m.factories
}

// Execute executes the provided function `fn` within the context of transactions from all registered factories.
// It creates transactions from all registered factories and coordinates them as a single logical transaction.
//
// If any transaction fails during commit, all transactions are rolled back to maintain consistency.
// The function `fn` receives a context that has all transactions set, allowing it to perform database operations
// across multiple transaction contexts.
func (m *TxManager) Execute(ctx context.Context, fn func(ctx context.Context) error) (err error) {
	if len(m.factories) == 0 {
		return errors.New("no transaction factories registered")
	}

	type txInfo struct {
		tx       Transaction
		executor TxDriver
	}

	transactions := make([]txInfo, 0, len(m.factories))
	txCtx := ctx

	// Create transactions from all factories
	for i := range m.factories {
		tx, err := m.factories[i].NewTx(ctx)
		if err != nil {
			// Rollback any already created transactions
			for _, txInfo := range transactions {
				_ = txInfo.tx.Rollback(ctx)
			}
			return fmt.Errorf("failed to create transaction for executor %s: %w", m.factories[i].Driver(), err)
		}
		transactions = append(transactions, txInfo{tx: tx, executor: m.factories[i].Driver()})
		txCtx = WithTxContext(txCtx, m.factories[i].Driver(), tx)
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

// --- Utilities ---

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
	txCtx := WithTxContext(ctx, factory.Driver(), tx)
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
