package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/stalkerxxl/soccer-team-api/internal/domain"
	"github.com/uptrace/bun"
)

var ErrTeamNotFound = errors.New("team not found")
var ErrCountryNotFound = errors.New("country not found")

// TeamRepository defines persistence operations for teams and countries.
type TeamRepository interface {
	Create(ctx context.Context, team *domain.Team) error
	GetByID(ctx context.Context, id int64) (*domain.Team, error)
	GetAll(ctx context.Context) ([]domain.Team, error)
	GetByUserID(ctx context.Context, userID int64) (*domain.Team, error)
	Update(ctx context.Context, team *domain.Team) error
	GetCountryByID(ctx context.Context, id int64) (*domain.Country, error)
	ListCountries(ctx context.Context) ([]domain.Country, error)
}

// TeamRepo is a Bun-backed implementation of TeamRepository.
type TeamRepo struct {
	db bun.IDB
}

// NewTeamRepository creates a team repository backed by the provided Bun handle.
func NewTeamRepository(db bun.IDB) *TeamRepo {
	return &TeamRepo{db: db}
}

// Create inserts a new team.
func (r *TeamRepo) Create(ctx context.Context, team *domain.Team) error {
	if _, err := r.db.NewInsert().Model(team).Exec(ctx); err != nil {
		return fmt.Errorf("create team: %w", err)
	}

	return nil
}

// GetByID returns a team by id with related country data preloaded.
func (r *TeamRepo) GetByID(ctx context.Context, id int64) (*domain.Team, error) {
	team := new(domain.Team)

	err := r.withTotalPlayersMarketValue(r.db.NewSelect().Model(team)).
		Relation("Country").
		Where("t.id = ?", id).
		Limit(1).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTeamNotFound
		}

		return nil, fmt.Errorf("get team by id: %w", err)
	}

	return team, nil
}

// GetByIDForUpdate locks and returns a team by id inside a transaction.
func (r *TeamRepo) GetByIDForUpdate(ctx context.Context, id int64) (*domain.Team, error) {
	team := new(domain.Team)

	err := r.db.NewSelect().
		Model(team).
		Where("t.id = ?", id).
		Limit(1).
		For("UPDATE").
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTeamNotFound
		}

		return nil, fmt.Errorf("get team by id for update: %w", err)
	}

	return team, nil
}

// GetAll returns all teams ordered by id.
func (r *TeamRepo) GetAll(ctx context.Context) ([]domain.Team, error) {
	var teams []domain.Team

	if err := r.db.NewSelect().
		Model(&teams).
		Relation("Country").
		Order("t.id ASC").
		Scan(ctx); err != nil {
		return nil, fmt.Errorf("get all teams: %w", err)
	}

	return teams, nil
}

// GetByUserID returns the team owned by the given user.
func (r *TeamRepo) GetByUserID(ctx context.Context, userID int64) (*domain.Team, error) {
	team := new(domain.Team)

	err := r.withTotalPlayersMarketValue(r.db.NewSelect().Model(team)).
		Relation("Country").
		Where("t.user_id = ?", userID).
		Limit(1).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTeamNotFound
		}

		return nil, fmt.Errorf("get team by user id: %w", err)
	}

	return team, nil
}

// Update persists changes to an existing team.
func (r *TeamRepo) Update(ctx context.Context, team *domain.Team) error {
	if _, err := r.db.NewUpdate().
		Model(team).
		WherePK().
		ExcludeColumn("created_at").
		Exec(ctx); err != nil {
		return fmt.Errorf("update team: %w", err)
	}

	return nil
}

// GetCountryByID returns a country by id.
func (r *TeamRepo) GetCountryByID(ctx context.Context, id int64) (*domain.Country, error) {
	country := new(domain.Country)

	err := r.db.NewSelect().
		Model(country).
		Where("c.id = ?", id).
		Limit(1).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCountryNotFound
		}

		return nil, fmt.Errorf("get country by id: %w", err)
	}

	return country, nil
}

// ListCountries returns all countries ordered by name.
func (r *TeamRepo) ListCountries(ctx context.Context) ([]domain.Country, error) {
	var countries []domain.Country

	if err := r.db.NewSelect().
		Model(&countries).
		Order("c.name ASC").
		Scan(ctx); err != nil {
		return nil, fmt.Errorf("list countries: %w", err)
	}

	return countries, nil
}

// withTotalPlayersMarketValue augments a team query with the computed squad market value.
func (r *TeamRepo) withTotalPlayersMarketValue(query *bun.SelectQuery) *bun.SelectQuery {
	return query.
		Column("t.*").
		ColumnExpr(
			"(SELECT COALESCE(SUM(p.market_value), 0) FROM players AS p WHERE p.team_id = t.id) AS total_players_market_value",
		)
}
