package service

import (
	"context"
	"errors"
	"strings"

	"github.com/stalkerxxl/soccer-team-api/internal/domain"
	"github.com/stalkerxxl/soccer-team-api/internal/repository"
)

var (
	// ErrTeamRepositoryNil indicates that the service requires a team repository.
	ErrTeamRepositoryNil = errors.New("team repository is required")
	// ErrEmptyTeamName indicates that the team name is required.
	ErrEmptyTeamName = errors.New("team name is required")
)

// TeamManager defines team use cases exposed to handlers and other services.
type TeamManager interface {
	CreateDefault(ctx context.Context, userID int64) (*domain.Team, error)
	GetByID(ctx context.Context, id int64) (*domain.Team, error)
	GetAll(ctx context.Context) ([]domain.Team, error)
	GetByUserID(ctx context.Context, userID int64) (*domain.Team, error)
	Update(ctx context.Context, team *domain.Team) error
	GetCountryByID(ctx context.Context, id int64) (*domain.Country, error)
	ListCountries(ctx context.Context) ([]domain.Country, error)
}

// TeamService implements team-related use cases.
type TeamService struct {
	teamRepo repository.TeamRepository
}

// NewTeamService creates a team service backed by the provided repository.
func NewTeamService(teamRepo repository.TeamRepository) (*TeamService, error) {
	if teamRepo == nil {
		return nil, ErrTeamRepositoryNil
	}

	return &TeamService{teamRepo: teamRepo}, nil
}

// CreateDefault creates the initial team for a newly registered user.
func (s *TeamService) CreateDefault(ctx context.Context, userID int64) (*domain.Team, error) {
	team := &domain.Team{
		UserID:    userID,
		Name:      domain.DefaultTeamName,
		CountryID: domain.DefaultCountryID,
		Budget:    domain.InitialTeamBudget,
	}

	if err := s.teamRepo.Create(ctx, team); err != nil {
		return nil, err
	}

	return team, nil
}

// GetByID returns a team by id.
func (s *TeamService) GetByID(ctx context.Context, id int64) (*domain.Team, error) {
	return s.teamRepo.GetByID(ctx, id)
}

// GetAll returns all teams.
func (s *TeamService) GetAll(ctx context.Context) ([]domain.Team, error) {
	return s.teamRepo.GetAll(ctx)
}

// GetByUserID returns the team owned by the given user.
func (s *TeamService) GetByUserID(ctx context.Context, userID int64) (*domain.Team, error) {
	return s.teamRepo.GetByUserID(ctx, userID)
}

// Update validates and persists team changes.
func (s *TeamService) Update(ctx context.Context, team *domain.Team) error {
	team.Name = strings.TrimSpace(team.Name)
	if team.Name == "" {
		return ErrEmptyTeamName
	}

	if _, err := s.teamRepo.GetCountryByID(ctx, team.CountryID); err != nil {
		return err
	}

	return s.teamRepo.Update(ctx, team)
}

// GetCountryByID returns a country by id.
func (s *TeamService) GetCountryByID(ctx context.Context, id int64) (*domain.Country, error) {
	return s.teamRepo.GetCountryByID(ctx, id)
}

// ListCountries returns all available countries.
func (s *TeamService) ListCountries(ctx context.Context) ([]domain.Country, error) {
	return s.teamRepo.ListCountries(ctx)
}
