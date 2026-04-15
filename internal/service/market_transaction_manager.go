package service

import (
	"context"
	"errors"

	dbtx "github.com/stalkerxxl/soccer-team-api/internal/db"
	"github.com/stalkerxxl/soccer-team-api/internal/repository"
	"github.com/uptrace/bun"
)

// ErrMarketTransactionManagerNil indicates that the market service requires a transaction manager.
var ErrMarketTransactionManagerNil = errors.New("market transaction manager is required")

// MarketTxDeps groups transaction-scoped repositories used during a market purchase.
type MarketTxDeps struct {
	ListingRepo  marketListingRepository
	PlayerRepo   marketPlayerRepository
	TeamRepo     marketTeamRepository
	TransferRepo marketTransferRepository
}

// MarketTransactionManager executes market workflows inside a transaction and injects tx-scoped dependencies.
type MarketTransactionManager interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context, deps MarketTxDeps) error) error
}

// marketTransactionManager is the default MarketTransactionManager implementation.
type marketTransactionManager struct {
	txManager dbtx.TxManager
}

// NewMarketTransactionManager creates a market transaction manager backed by the provided DB transaction manager.
func NewMarketTransactionManager(txManager dbtx.TxManager) (MarketTransactionManager, error) {
	if txManager == nil {
		return nil, ErrMarketTransactionManagerNil
	}

	return &marketTransactionManager{txManager: txManager}, nil
}

// WithinTransaction runs fn inside a database transaction with repositories bound to txDB.
func (m *marketTransactionManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context, deps MarketTxDeps) error) error {
	return m.txManager.WithinTransaction(ctx, func(ctx context.Context, txDB bun.IDB) error {
		// Each repository must be created on top of txDB so all writes participate in the same transaction.
		return fn(ctx, MarketTxDeps{
			ListingRepo:  repository.NewMarketListingRepository(txDB),
			PlayerRepo:   repository.NewPlayerRepository(txDB),
			TeamRepo:     repository.NewTeamRepository(txDB),
			TransferRepo: repository.NewTransferRepository(txDB),
		})
	})
}
