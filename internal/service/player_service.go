package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/stalkerxxl/soccer-team-api/internal/domain"
	"github.com/stalkerxxl/soccer-team-api/internal/repository"
)

var (
	// ErrPlayerRepositoryNil indicates that the service requires a player repository.
	ErrPlayerRepositoryNil = errors.New("player repository is required")
	// ErrPlayerCountryProviderNil indicates that the service requires a country provider.
	ErrPlayerCountryProviderNil = errors.New("player country provider is required")
	// ErrPlayerGeneratorNil indicates that the service requires a player data generator.
	ErrPlayerGeneratorNil = errors.New("player data generator is required")
	// ErrEmptyPlayerFirstName indicates that the player first name is required.
	ErrEmptyPlayerFirstName = errors.New("player first name is required")
	// ErrEmptyPlayerLastName indicates that the player last name is required.
	ErrEmptyPlayerLastName = errors.New("player last name is required")
	// ErrNoCountriesAvailable indicates that player generation cannot proceed without countries.
	ErrNoCountriesAvailable = errors.New("no countries available for player generation")
	// ErrInvalidInitialSquadConfig indicates that the starter squad composition is inconsistent.
	ErrInvalidInitialSquadConfig = errors.New("initial squad composition is invalid")
	// ErrInvalidGeneratedPlayerData indicates that generated player data failed validation.
	ErrInvalidGeneratedPlayerData = errors.New("generated player data is invalid")
)

// PlayerCountryProvider supplies country data required by player workflows.
type PlayerCountryProvider interface {
	GetCountryByID(ctx context.Context, id int64) (*domain.Country, error)
	ListCountries(ctx context.Context) ([]domain.Country, error)
}

// PlayerManager defines player use cases exposed to other services and handlers.
type PlayerManager interface {
	CreateInitialSquad(ctx context.Context, teamID int64) ([]domain.Player, error)
	GetByID(ctx context.Context, id int64) (*domain.Player, error)
	ListByTeamID(ctx context.Context, teamID int64) ([]domain.Player, error)
	Update(ctx context.Context, player *domain.Player) error
}

// PlayerService implements player-related use cases.
type PlayerService struct {
	playerRepo      repository.PlayerRepository
	countryProvider PlayerCountryProvider
	playerGenerator playerDataGenerator
}

// generatedPlayerData is the raw data produced by the player generator before persistence.
type generatedPlayerData struct {
	FirstName string
	LastName  string
	CountryID int64
	Age       int
}

// playerDataGenerator produces randomized player attributes for initial squad generation.
type playerDataGenerator interface {
	Generate(countries []domain.Country) (generatedPlayerData, error)
}

// randomPlayerDataGenerator provides a concurrency-safe random player data generator.
type randomPlayerDataGenerator struct {
	mu  sync.Mutex
	rng *rand.Rand
}

var (
	// defaultPlayerFirstNames is the fallback pool used for generated first names.
	defaultPlayerFirstNames = []string{
		"Liam", "Noah", "Oliver", "Ethan", "Mason", "Lucas", "Leo", "Mateo",
		"Santiago", "Julian", "Theo", "Felix", "Daniel", "Adrian", "Nico", "Victor",
		"Marco", "Ruben", "Gabriel", "Mateus", "Bruno", "Tomas", "Samuel", "David",
	}
	// defaultPlayerLastNames is the fallback pool used for generated last names.
	defaultPlayerLastNames = []string{
		"Smith", "Brown", "Taylor", "Wilson", "Evans", "Bennett", "Silva", "Costa",
		"Fernandes", "Rossi", "Martin", "Garcia", "Lopez", "Pereira", "Alonso", "Muller",
		"Schneider", "Keller", "van Dijk", "de Jong", "Santos", "Rivera", "Torres", "Navarro",
	}
)

// NewPlayerService creates a player service with runtime randomness enabled.
func NewPlayerService(
	playerRepo repository.PlayerRepository,
	countryProvider PlayerCountryProvider,
) (*PlayerService, error) {
	generator := &randomPlayerDataGenerator{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	return newPlayerService(playerRepo, countryProvider, generator)
}

// newPlayerService keeps the public constructor small while letting tests inject
// a deterministic generator instead of runtime randomness.
func newPlayerService(
	playerRepo repository.PlayerRepository,
	countryProvider PlayerCountryProvider,
	generator playerDataGenerator,
) (*PlayerService, error) {
	if playerRepo == nil {
		return nil, ErrPlayerRepositoryNil
	}

	if countryProvider == nil {
		return nil, ErrPlayerCountryProviderNil
	}

	if generator == nil {
		return nil, ErrPlayerGeneratorNil
	}

	return &PlayerService{
		playerRepo:      playerRepo,
		countryProvider: countryProvider,
		playerGenerator: generator,
	}, nil
}

// CreateInitialSquad generates and persists the starter squad for a newly created team.
func (s *PlayerService) CreateInitialSquad(ctx context.Context, teamID int64) ([]domain.Player, error) {
	countries, err := s.countryProvider.ListCountries(ctx)
	if err != nil {
		return nil, err
	}

	if len(countries) == 0 {
		return nil, ErrNoCountriesAvailable
	}

	players := make([]domain.Player, 0, domain.InitialSquadSize)

	for _, position := range domain.ValidPlayerPositions {
		count, ok := domain.InitialSquadComposition[position]
		if !ok || count < 1 {
			return nil, fmt.Errorf("%w: missing count for position %q", ErrInvalidInitialSquadConfig, position)
		}

		for range count {
			playerData, err := s.playerGenerator.Generate(countries)
			if err != nil {
				return nil, err
			}

			if err := validateGeneratedPlayerData(playerData); err != nil {
				return nil, err
			}

			players = append(players, domain.Player{
				TeamID:      teamID,
				FirstName:   playerData.FirstName,
				LastName:    playerData.LastName,
				CountryID:   playerData.CountryID,
				Age:         playerData.Age,
				Position:    position,
				MarketValue: domain.InitialPlayerMarketValue,
			})
		}
	}

	if len(players) != domain.InitialSquadSize {
		return nil, fmt.Errorf("%w: got %d players, want %d", ErrInvalidInitialSquadConfig, len(players), domain.InitialSquadSize)
	}

	if err := s.playerRepo.CreateBulk(ctx, players); err != nil {
		return nil, err
	}

	return players, nil
}

// GetByID returns a player by id.
func (s *PlayerService) GetByID(ctx context.Context, id int64) (*domain.Player, error) {
	return s.playerRepo.GetByID(ctx, id)
}

// ListByTeamID returns all players that belong to the given team.
func (s *PlayerService) ListByTeamID(ctx context.Context, teamID int64) ([]domain.Player, error) {
	return s.playerRepo.ListByTeamID(ctx, teamID)
}

// Update validates and persists player changes.
func (s *PlayerService) Update(ctx context.Context, player *domain.Player) error {
	player.FirstName = strings.TrimSpace(player.FirstName)
	player.LastName = strings.TrimSpace(player.LastName)

	if player.FirstName == "" {
		return ErrEmptyPlayerFirstName
	}

	if player.LastName == "" {
		return ErrEmptyPlayerLastName
	}

	if _, err := s.countryProvider.GetCountryByID(ctx, player.CountryID); err != nil {
		return err
	}

	return s.playerRepo.Update(ctx, player)
}

// Generate returns randomized player data using the configured countries pool.
func (g *randomPlayerDataGenerator) Generate(countries []domain.Country) (generatedPlayerData, error) {
	if len(countries) == 0 {
		return generatedPlayerData{}, ErrNoCountriesAvailable
	}

	// rand.Rand is not safe for concurrent use, so generation must be serialized.
	g.mu.Lock()
	defer g.mu.Unlock()

	country := countries[g.rng.Intn(len(countries))]
	ageRange := domain.MaxPlayerAge - domain.MinPlayerAge + 1

	return generatedPlayerData{
		FirstName: defaultPlayerFirstNames[g.rng.Intn(len(defaultPlayerFirstNames))],
		LastName:  defaultPlayerLastNames[g.rng.Intn(len(defaultPlayerLastNames))],
		CountryID: country.ID,
		Age:       domain.MinPlayerAge + g.rng.Intn(ageRange),
	}, nil
}

// validateGeneratedPlayerData checks that generator output satisfies domain constraints.
func validateGeneratedPlayerData(playerData generatedPlayerData) error {
	if strings.TrimSpace(playerData.FirstName) == "" {
		return fmt.Errorf("%w: first name", ErrInvalidGeneratedPlayerData)
	}

	if strings.TrimSpace(playerData.LastName) == "" {
		return fmt.Errorf("%w: last name", ErrInvalidGeneratedPlayerData)
	}

	if playerData.CountryID < 1 {
		return fmt.Errorf("%w: country id", ErrInvalidGeneratedPlayerData)
	}

	if playerData.Age < domain.MinPlayerAge || playerData.Age > domain.MaxPlayerAge {
		return fmt.Errorf("%w: age %d", ErrInvalidGeneratedPlayerData, playerData.Age)
	}

	return nil
}
