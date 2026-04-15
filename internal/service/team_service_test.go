package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stalkerxxl/soccer-team-api/internal/domain"
	"github.com/stalkerxxl/soccer-team-api/internal/repository"
)

func TestNewTeamServiceRequiresRepository(t *testing.T) {
	t.Parallel()

	_, err := NewTeamService(nil)
	if !errors.Is(err, ErrTeamRepositoryNil) {
		t.Fatalf("NewTeamService() error = %v, want %v", err, ErrTeamRepositoryNil)
	}
}

func TestTeamServiceCreateDefault(t *testing.T) {
	t.Parallel()

	repo := &stubTeamRepository{}
	svc := newTestTeamService(t, repo)

	team, err := svc.CreateDefault(context.Background(), 42)
	if err != nil {
		t.Fatalf("CreateDefault() error = %v", err)
	}

	if team == nil {
		t.Fatal("CreateDefault() returned nil team")
	}

	if repo.createdTeam == nil {
		t.Fatal("CreateDefault() did not call repository Create")
	}

	if team.UserID != 42 {
		t.Fatalf("team.UserID = %d, want %d", team.UserID, 42)
	}

	if team.Name != domain.DefaultTeamName {
		t.Fatalf("team.Name = %q, want %q", team.Name, domain.DefaultTeamName)
	}

	if team.CountryID != domain.DefaultCountryID {
		t.Fatalf("team.CountryID = %d, want %d", team.CountryID, domain.DefaultCountryID)
	}

	if team.Budget != domain.InitialTeamBudget {
		t.Fatalf("team.Budget = %d, want %d", team.Budget, domain.InitialTeamBudget)
	}
}

func TestTeamServiceUpdateRejectsEmptyName(t *testing.T) {
	t.Parallel()

	repo := &stubTeamRepository{}
	svc := newTestTeamService(t, repo)

	err := svc.Update(context.Background(), &domain.Team{
		ID:        1,
		UserID:    7,
		Name:      "   ",
		CountryID: domain.DefaultCountryID,
	})
	if !errors.Is(err, ErrEmptyTeamName) {
		t.Fatalf("Update() error = %v, want %v", err, ErrEmptyTeamName)
	}
}

func TestTeamServiceUpdateValidatesCountryAndNormalizesName(t *testing.T) {
	t.Parallel()

	repo := &stubTeamRepository{
		countryByID: &domain.Country{ID: 2, Name: "Georgia", ISOCode: "GE"},
	}
	svc := newTestTeamService(t, repo)

	team := &domain.Team{
		ID:        1,
		UserID:    7,
		Name:      "  Dinamo Tbilisi  ",
		CountryID: 2,
		Budget:    domain.InitialTeamBudget,
	}

	if err := svc.Update(context.Background(), team); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	if repo.updatedTeam == nil {
		t.Fatal("Update() did not call repository Update")
	}

	if team.Name != "Dinamo Tbilisi" {
		t.Fatalf("team.Name = %q, want %q", team.Name, "Dinamo Tbilisi")
	}

	if repo.lastCountryID != 2 {
		t.Fatalf("GetCountryByID() id = %d, want %d", repo.lastCountryID, 2)
	}
}

func TestTeamServiceUpdateReturnsCountryError(t *testing.T) {
	t.Parallel()

	repo := &stubTeamRepository{
		getCountryErr: repository.ErrCountryNotFound,
	}
	svc := newTestTeamService(t, repo)

	err := svc.Update(context.Background(), &domain.Team{
		ID:        1,
		UserID:    7,
		Name:      "Valid name",
		CountryID: 999,
	})
	if !errors.Is(err, repository.ErrCountryNotFound) {
		t.Fatalf("Update() error = %v, want %v", err, repository.ErrCountryNotFound)
	}
}

func TestTeamServiceGetByUserIDPopulatesTotalPlayersMarketValue(t *testing.T) {
	t.Parallel()

	repo := &stubTeamRepository{
		teamByUserID: &domain.Team{ID: 10, UserID: 42, Name: "Your team", TotalPlayersMarketValue: 7_500_000},
	}
	svc := newTestTeamService(t, repo)

	team, err := svc.GetByUserID(context.Background(), 42)
	if err != nil {
		t.Fatalf("GetByUserID() error = %v", err)
	}

	if team.TotalPlayersMarketValue != 7_500_000 {
		t.Fatalf("team.TotalPlayersMarketValue = %d, want %d", team.TotalPlayersMarketValue, int64(7_500_000))
	}
}

func TestTeamServiceGetByIDPopulatesTotalPlayersMarketValue(t *testing.T) {
	t.Parallel()

	repo := &stubTeamRepository{
		teamByID: &domain.Team{ID: 12, UserID: 7, Name: "Dinamo", TotalPlayersMarketValue: 8_000_000},
	}
	svc := newTestTeamService(t, repo)

	team, err := svc.GetByID(context.Background(), 12)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if team.TotalPlayersMarketValue != 8_000_000 {
		t.Fatalf("team.TotalPlayersMarketValue = %d, want %d", team.TotalPlayersMarketValue, int64(8_000_000))
	}
}

func newTestTeamService(t *testing.T, teamRepo *stubTeamRepository) TeamManager {
	t.Helper()

	svc, err := NewTeamService(teamRepo)
	if err != nil {
		t.Fatalf("NewTeamService() error = %v", err)
	}

	return svc
}

type stubTeamRepository struct {
	teamByID       *domain.Team
	teamByUserID   *domain.Team
	teams          []domain.Team
	countryByID    *domain.Country
	countries      []domain.Country
	createdTeam    *domain.Team
	updatedTeam    *domain.Team
	lastCountryID  int64
	createErr      error
	getByIDErr     error
	getAllErr      error
	getByUserIDErr error
	updateErr      error
	getCountryErr  error
	listErr        error
}

func (r *stubTeamRepository) Create(_ context.Context, team *domain.Team) error {
	r.createdTeam = team
	return r.createErr
}

func (r *stubTeamRepository) GetByID(_ context.Context, _ int64) (*domain.Team, error) {
	return r.teamByID, r.getByIDErr
}

func (r *stubTeamRepository) GetAll(_ context.Context) ([]domain.Team, error) {
	return r.teams, r.getAllErr
}

func (r *stubTeamRepository) GetByUserID(_ context.Context, _ int64) (*domain.Team, error) {
	return r.teamByUserID, r.getByUserIDErr
}

func (r *stubTeamRepository) Update(_ context.Context, team *domain.Team) error {
	r.updatedTeam = team
	return r.updateErr
}

func (r *stubTeamRepository) GetCountryByID(_ context.Context, id int64) (*domain.Country, error) {
	r.lastCountryID = id
	return r.countryByID, r.getCountryErr
}

func (r *stubTeamRepository) ListCountries(_ context.Context) ([]domain.Country, error) {
	return r.countries, r.listErr
}
