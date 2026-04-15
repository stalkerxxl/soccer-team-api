package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/stalkerxxl/soccer-team-api/internal/domain"
	"github.com/uptrace/bun"
)

var ErrPlayerNotFound = errors.New("player not found")

// PlayerRepository defines persistence operations for players.
type PlayerRepository interface {
	Create(ctx context.Context, player *domain.Player) error
	CreateBulk(ctx context.Context, players []domain.Player) error
	GetByID(ctx context.Context, id int64) (*domain.Player, error)
	ListByTeamID(ctx context.Context, teamID int64) ([]domain.Player, error)
	Update(ctx context.Context, player *domain.Player) error
}

// PlayerRepo is a Bun-backed implementation of PlayerRepository.
type PlayerRepo struct {
	db bun.IDB
}

// NewPlayerRepository creates a player repository backed by the provided Bun handle.
func NewPlayerRepository(db bun.IDB) *PlayerRepo {
	return &PlayerRepo{db: db}
}

// Create inserts a new player.
func (r *PlayerRepo) Create(ctx context.Context, player *domain.Player) error {
	if _, err := r.db.NewInsert().Model(player).Exec(ctx); err != nil {
		return fmt.Errorf("create player: %w", err)
	}

	return nil
}

// CreateBulk inserts multiple players in a single query.
func (r *PlayerRepo) CreateBulk(ctx context.Context, players []domain.Player) error {
	if len(players) == 0 {
		return nil
	}

	if _, err := r.db.NewInsert().Model(&players).Exec(ctx); err != nil {
		return fmt.Errorf("create players bulk: %w", err)
	}

	return nil
}

// GetByID returns a player by id with related country data preloaded.
func (r *PlayerRepo) GetByID(ctx context.Context, id int64) (*domain.Player, error) {
	player := new(domain.Player)

	err := r.db.NewSelect().
		Model(player).
		Relation("Country").
		Where("p.id = ?", id).
		Limit(1).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPlayerNotFound
		}

		return nil, fmt.Errorf("get player by id: %w", err)
	}

	return player, nil
}

// GetByIDForUpdate locks and returns a player by id inside a transaction.
func (r *PlayerRepo) GetByIDForUpdate(ctx context.Context, id int64) (*domain.Player, error) {
	player := new(domain.Player)

	err := r.db.NewSelect().
		Model(player).
		Where("p.id = ?", id).
		Limit(1).
		For("UPDATE").
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPlayerNotFound
		}

		return nil, fmt.Errorf("get player by id for update: %w", err)
	}

	return player, nil
}

// ListByTeamID returns all players that belong to the given team.
func (r *PlayerRepo) ListByTeamID(ctx context.Context, teamID int64) ([]domain.Player, error) {
	var players []domain.Player

	if err := r.db.NewSelect().
		Model(&players).
		Relation("Country").
		Where("p.team_id = ?", teamID).
		Order("p.id ASC").
		Scan(ctx); err != nil {
		return nil, fmt.Errorf("list players by team id: %w", err)
	}

	return players, nil
}

// Update persists changes to an existing player.
func (r *PlayerRepo) Update(ctx context.Context, player *domain.Player) error {
	if _, err := r.db.NewUpdate().
		Model(player).
		WherePK().
		ExcludeColumn("created_at").
		Exec(ctx); err != nil {
		return fmt.Errorf("update player: %w", err)
	}

	return nil
}
