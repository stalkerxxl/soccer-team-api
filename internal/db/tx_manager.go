package db

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

// TxManager executes callbacks inside a database transaction.
type TxManager interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context, txDB bun.IDB) error) error
}

// BunTxManager implements TxManager on top of bun.DB.
type BunTxManager struct {
	db *bun.DB
}

// NewTxManager creates a transaction manager backed by the provided Bun database handle.
func NewTxManager(db *bun.DB) *BunTxManager {
	return &BunTxManager{db: db}
}

// WithinTransaction runs fn inside a Bun transaction and forwards the transactional handle.
func (m *BunTxManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context, txDB bun.IDB) error) error {
	if err := m.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		return fn(ctx, tx)
	}); err != nil {
		return fmt.Errorf("run transaction: %w", err)
	}

	return nil
}
