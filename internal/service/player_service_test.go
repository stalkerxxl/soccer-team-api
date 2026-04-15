package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stalkerxxl/soccer-team-api/internal/domain"
	"github.com/stalkerxxl/soccer-team-api/internal/repository"
)

func TestNewPlayerServiceRequiresDependencies(t *testing.T) {
	t.Parallel()

	_, err := newPlayerService(nil, &stubPlayerCountryProvider{}, &stubPlayerDataGenerator{})
	if !errors.Is(err, ErrPlayerRepositoryNil) {
		t.Fatalf("newPlayerService() repo error = %v, want %v", err, ErrPlayerRepositoryNil)
	}

	_, err = newPlayerService(&stubPlayerRepository{}, nil, &stubPlayerDataGenerator{})
	if !errors.Is(err, ErrPlayerCountryProviderNil) {
		t.Fatalf("newPlayerService() country provider error = %v, want %v", err, ErrPlayerCountryProviderNil)
	}

	_, err = newPlayerService(&stubPlayerRepository{}, &stubPlayerCountryProvider{}, nil)
	if !errors.Is(err, ErrPlayerGeneratorNil) {
		t.Fatalf("newPlayerService() generator error = %v, want %v", err, ErrPlayerGeneratorNil)
	}
}

func TestPlayerServiceCreateInitialSquad(t *testing.T) {
	t.Parallel()

	playerRepo := &stubPlayerRepository{}
	countryProvider := &stubPlayerCountryProvider{
		countries: []domain.Country{
			{ID: 1, Name: "United Kingdom", ISOCode: "GB"},
			{ID: 2, Name: "Georgia", ISOCode: "GE"},
		},
	}
	generator := &stubPlayerDataGenerator{
		sequence: []generatedPlayerData{
			{FirstName: "Liam", LastName: "Smith", CountryID: 1, Age: 21},
			{FirstName: "Noah", LastName: "Brown", CountryID: 2, Age: 29},
		},
	}
	svc := newTestPlayerService(t, playerRepo, countryProvider, generator)

	players, err := svc.CreateInitialSquad(context.Background(), 99)
	if err != nil {
		t.Fatalf("CreateInitialSquad() error = %v", err)
	}

	if len(players) != domain.InitialSquadSize {
		t.Fatalf("len(players) = %d, want %d", len(players), domain.InitialSquadSize)
	}

	if playerRepo.createdPlayers == nil {
		t.Fatal("CreateInitialSquad() did not call repository CreateBulk")
	}

	if len(playerRepo.createdPlayers) != domain.InitialSquadSize {
		t.Fatalf("len(createdPlayers) = %d, want %d", len(playerRepo.createdPlayers), domain.InitialSquadSize)
	}

	positionCounts := make(map[domain.PlayerPosition]int)

	for _, player := range players {
		if player.TeamID != 99 {
			t.Fatalf("player.TeamID = %d, want %d", player.TeamID, 99)
		}

		if player.MarketValue != domain.InitialPlayerMarketValue {
			t.Fatalf("player.MarketValue = %d, want %d", player.MarketValue, domain.InitialPlayerMarketValue)
		}

		if player.Age < domain.MinPlayerAge || player.Age > domain.MaxPlayerAge {
			t.Fatalf("player.Age = %d, want between %d and %d", player.Age, domain.MinPlayerAge, domain.MaxPlayerAge)
		}

		positionCounts[player.Position]++
	}

	for position, wantCount := range domain.InitialSquadComposition {
		if got := positionCounts[position]; got != wantCount {
			t.Fatalf("position %q count = %d, want %d", position, got, wantCount)
		}
	}
}

func TestPlayerServiceCreateInitialSquadReturnsErrorWhenCountriesMissing(t *testing.T) {
	t.Parallel()

	svc := newTestPlayerService(
		t,
		&stubPlayerRepository{},
		&stubPlayerCountryProvider{},
		&stubPlayerDataGenerator{},
	)

	_, err := svc.CreateInitialSquad(context.Background(), 1)
	if !errors.Is(err, ErrNoCountriesAvailable) {
		t.Fatalf("CreateInitialSquad() error = %v, want %v", err, ErrNoCountriesAvailable)
	}
}

func TestPlayerServiceUpdateValidatesFields(t *testing.T) {
	t.Parallel()

	playerRepo := &stubPlayerRepository{}
	countryProvider := &stubPlayerCountryProvider{
		countryByID: &domain.Country{ID: 2, Name: "Georgia", ISOCode: "GE"},
	}
	svc := newTestPlayerService(t, playerRepo, countryProvider, &stubPlayerDataGenerator{})

	player := &domain.Player{
		ID:        7,
		TeamID:    5,
		FirstName: "  Nico ",
		LastName:  " Williams  ",
		CountryID: 2,
		Position:  domain.PlayerPositionAttacker,
		Age:       23,
	}

	if err := svc.Update(context.Background(), player); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	if player.FirstName != "Nico" {
		t.Fatalf("player.FirstName = %q, want %q", player.FirstName, "Nico")
	}

	if player.LastName != "Williams" {
		t.Fatalf("player.LastName = %q, want %q", player.LastName, "Williams")
	}

	if playerRepo.updatedPlayer == nil {
		t.Fatal("Update() did not call repository Update")
	}
}

func TestPlayerServiceUpdateReturnsCountryError(t *testing.T) {
	t.Parallel()

	svc := newTestPlayerService(
		t,
		&stubPlayerRepository{},
		&stubPlayerCountryProvider{getCountryErr: repository.ErrCountryNotFound},
		&stubPlayerDataGenerator{},
	)

	err := svc.Update(context.Background(), &domain.Player{
		ID:        1,
		TeamID:    1,
		FirstName: "Liam",
		LastName:  "Smith",
		CountryID: 999,
	})
	if !errors.Is(err, repository.ErrCountryNotFound) {
		t.Fatalf("Update() error = %v, want %v", err, repository.ErrCountryNotFound)
	}
}

func newTestPlayerService(
	t *testing.T,
	playerRepo *stubPlayerRepository,
	countryProvider *stubPlayerCountryProvider,
	generator *stubPlayerDataGenerator,
) PlayerManager {
	t.Helper()

	svc, err := newPlayerService(playerRepo, countryProvider, generator)
	if err != nil {
		t.Fatalf("newPlayerService() error = %v", err)
	}

	return svc
}

type stubPlayerRepository struct {
	playerByID     *domain.Player
	playersByTeam  []domain.Player
	createdPlayers []domain.Player
	updatedPlayer  *domain.Player
	lastPlayerID   int64
	createErr      error
	getByIDErr     error
	listErr        error
	updateErr      error
}

func (r *stubPlayerRepository) Create(_ context.Context, _ *domain.Player) error {
	return nil
}

func (r *stubPlayerRepository) CreateBulk(_ context.Context, players []domain.Player) error {
	r.createdPlayers = append([]domain.Player(nil), players...)
	return r.createErr
}

func (r *stubPlayerRepository) GetByID(_ context.Context, id int64) (*domain.Player, error) {
	r.lastPlayerID = id
	return r.playerByID, r.getByIDErr
}

func (r *stubPlayerRepository) ListByTeamID(_ context.Context, _ int64) ([]domain.Player, error) {
	return r.playersByTeam, r.listErr
}

func (r *stubPlayerRepository) Update(_ context.Context, player *domain.Player) error {
	r.updatedPlayer = player
	return r.updateErr
}

type stubPlayerCountryProvider struct {
	countryByID   *domain.Country
	countries     []domain.Country
	lastCountryID int64
	getCountryErr error
	listErr       error
}

func (r *stubPlayerCountryProvider) GetCountryByID(_ context.Context, id int64) (*domain.Country, error) {
	r.lastCountryID = id
	return r.countryByID, r.getCountryErr
}

func (r *stubPlayerCountryProvider) ListCountries(_ context.Context) ([]domain.Country, error) {
	return r.countries, r.listErr
}

type stubPlayerDataGenerator struct {
	sequence []generatedPlayerData
	next     int
	err      error
}

func (g *stubPlayerDataGenerator) Generate(_ []domain.Country) (generatedPlayerData, error) {
	if g.err != nil {
		return generatedPlayerData{}, g.err
	}

	if len(g.sequence) == 0 {
		return generatedPlayerData{
			FirstName: "Liam",
			LastName:  "Smith",
			CountryID: 1,
			Age:       21,
		}, nil
	}

	playerData := g.sequence[g.next%len(g.sequence)]
	g.next++

	return playerData, nil
}
