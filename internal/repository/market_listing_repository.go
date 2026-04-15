package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/stalkerxxl/soccer-team-api/internal/domain"
	"github.com/uptrace/bun"
)

var ErrMarketListingNotFound = errors.New("market listing not found")

// MarketListingRepository defines persistence operations for transfer market listings.
type MarketListingRepository interface {
	Create(ctx context.Context, listing *domain.MarketListing) error
	GetByID(ctx context.Context, id int64) (*domain.MarketListing, error)
	GetActiveByPlayerID(ctx context.Context, playerID int64) (*domain.MarketListing, error)
	ListActive(ctx context.Context) ([]domain.MarketListing, error)
	Update(ctx context.Context, listing *domain.MarketListing) error
}

// MarketListingRepo is a Bun-backed implementation of MarketListingRepository.
type MarketListingRepo struct {
	db bun.IDB
}

// NewMarketListingRepository creates a market listing repository backed by the provided Bun handle.
func NewMarketListingRepository(db bun.IDB) *MarketListingRepo {
	return &MarketListingRepo{db: db}
}

// Create inserts a new market listing.
func (r *MarketListingRepo) Create(ctx context.Context, listing *domain.MarketListing) error {
	if _, err := r.db.NewInsert().Model(listing).Exec(ctx); err != nil {
		return fmt.Errorf("create market listing: %w", err)
	}

	return nil
}

// GetByID returns a market listing by its identifier with related entities preloaded.
func (r *MarketListingRepo) GetByID(ctx context.Context, id int64) (*domain.MarketListing, error) {
	listing := new(domain.MarketListing)

	err := r.withRelations(r.db.NewSelect().Model(listing)).
		Where("ml.id = ?", id).
		Limit(1).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrMarketListingNotFound
		}

		return nil, fmt.Errorf("get market listing by id: %w", err)
	}

	return listing, nil
}

// GetActiveByPlayerID returns the active listing for a player.
func (r *MarketListingRepo) GetActiveByPlayerID(ctx context.Context, playerID int64) (*domain.MarketListing, error) {
	listing := new(domain.MarketListing)

	err := r.withRelations(r.db.NewSelect().Model(listing)).
		Where("ml.player_id = ?", playerID).
		Where("ml.status = ?", domain.MarketListingStatusActive).
		Limit(1).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrMarketListingNotFound
		}

		return nil, fmt.Errorf("get active market listing by player id: %w", err)
	}

	return listing, nil
}

// GetActiveByIDForUpdate locks and returns an active listing by id inside a transaction.
func (r *MarketListingRepo) GetActiveByIDForUpdate(ctx context.Context, id int64) (*domain.MarketListing, error) {
	listing := new(domain.MarketListing)

	err := r.db.NewSelect().
		Model(listing).
		Where("ml.id = ?", id).
		Where("ml.status = ?", domain.MarketListingStatusActive).
		Limit(1).
		For("UPDATE").
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrMarketListingNotFound
		}

		return nil, fmt.Errorf("get active market listing by id for update: %w", err)
	}

	return listing, nil
}

// ListActive returns all active listings ordered by newest first.
func (r *MarketListingRepo) ListActive(ctx context.Context) ([]domain.MarketListing, error) {
	var listings []domain.MarketListing

	if err := r.withRelations(r.db.NewSelect().Model(&listings)).
		Where("ml.status = ?", domain.MarketListingStatusActive).
		Order("ml.created_at DESC").
		Scan(ctx); err != nil {
		return nil, fmt.Errorf("list active market listings: %w", err)
	}

	return listings, nil
}

// Update persists changes to an existing market listing.
func (r *MarketListingRepo) Update(ctx context.Context, listing *domain.MarketListing) error {
	if _, err := r.db.NewUpdate().
		Model(listing).
		WherePK().
		ExcludeColumn("created_at").
		Exec(ctx); err != nil {
		return fmt.Errorf("update market listing: %w", err)
	}

	return nil
}

// withRelations keeps public market listing reads hydrated with the data expected by API responses.
func (r *MarketListingRepo) withRelations(query *bun.SelectQuery) *bun.SelectQuery {
	return query.
		Relation("Player").
		Relation("Player.Country").
		Relation("SellerTeam")
}
