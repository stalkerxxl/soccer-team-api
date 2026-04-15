package repository

import (
	"context"
	"fmt"

	"github.com/stalkerxxl/soccer-team-api/internal/domain"
	"github.com/uptrace/bun"
)

// TransferRepository defines persistence operations for completed transfers.
type TransferRepository interface {
	Create(ctx context.Context, transfer *domain.Transfer) error
}

// TransferRepo is a Bun-backed implementation of TransferRepository.
type TransferRepo struct {
	db bun.IDB
}

// NewTransferRepository creates a transfer repository backed by the provided Bun handle.
func NewTransferRepository(db bun.IDB) *TransferRepo {
	return &TransferRepo{db: db}
}

// Create inserts a completed transfer record.
func (r *TransferRepo) Create(ctx context.Context, transfer *domain.Transfer) error {
	if _, err := r.db.NewInsert().Model(transfer).Exec(ctx); err != nil {
		return fmt.Errorf("create transfer: %w", err)
	}

	return nil
}
