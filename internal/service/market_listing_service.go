package service

import (
	"context"
	"errors"
	"math/rand"
	"sync"

	"github.com/stalkerxxl/soccer-team-api/internal/domain"
	"github.com/stalkerxxl/soccer-team-api/internal/repository"
)

var (
	// ErrMarketListingRepositoryNil indicates that the service requires a listing repository.
	ErrMarketListingRepositoryNil = errors.New("market listing repository is required")
	// ErrMarketPlayerRepositoryNil indicates that the service requires a player repository.
	ErrMarketPlayerRepositoryNil = errors.New("player repository is required")
	// ErrInvalidAskingPrice indicates that a listing price must be positive.
	ErrInvalidAskingPrice = errors.New("asking price must be greater than zero")
	// ErrListingAlreadyActive indicates that the player is already listed on the market.
	ErrListingAlreadyActive = errors.New("market listing is already active")
	// ErrPlayerNotOwned indicates that the player does not belong to the acting team.
	ErrPlayerNotOwned = errors.New("player does not belong to current team")
	// ErrCannotBuyOwnPlayer indicates that a team cannot buy a player from itself.
	ErrCannotBuyOwnPlayer = errors.New("cannot buy own player")
	// ErrInsufficientBudget indicates that the buyer team cannot afford the asking price.
	ErrInsufficientBudget = errors.New("insufficient budget")
)

// marketListingRepository defines the listing persistence operations required by the service.
type marketListingRepository interface {
	Create(ctx context.Context, listing *domain.MarketListing) error
	GetByID(ctx context.Context, id int64) (*domain.MarketListing, error)
	GetActiveByPlayerID(ctx context.Context, playerID int64) (*domain.MarketListing, error)
	GetActiveByIDForUpdate(ctx context.Context, id int64) (*domain.MarketListing, error)
	ListActive(ctx context.Context) ([]domain.MarketListing, error)
	Update(ctx context.Context, listing *domain.MarketListing) error
}

// marketPlayerRepository defines the player persistence operations required by the service.
type marketPlayerRepository interface {
	GetByID(ctx context.Context, id int64) (*domain.Player, error)
	GetByIDForUpdate(ctx context.Context, id int64) (*domain.Player, error)
	Update(ctx context.Context, player *domain.Player) error
}

// marketTeamRepository defines the team persistence operations required by the buy flow.
type marketTeamRepository interface {
	GetByIDForUpdate(ctx context.Context, id int64) (*domain.Team, error)
	Update(ctx context.Context, team *domain.Team) error
}

// marketTransferRepository defines the transfer persistence operations required by the buy flow.
type marketTransferRepository interface {
	Create(ctx context.Context, transfer *domain.Transfer) error
}

// MarketListingManager defines transfer market use cases exposed to handlers.
type MarketListingManager interface {
	ListActive(ctx context.Context) ([]domain.MarketListing, error)
	CreateListing(ctx context.Context, sellerTeamID, playerID, askingPrice int64) (*domain.MarketListing, error)
	CancelListing(ctx context.Context, sellerTeamID, playerID int64) (*domain.MarketListing, error)
	BuyListing(ctx context.Context, buyerTeamID, listingID int64) (*domain.Transfer, error)
}

// MarketListingService implements transfer market use cases.
type MarketListingService struct {
	listingRepo       marketListingRepository
	playerRepo        marketPlayerRepository
	txManager         MarketTransactionManager
	valueIncreaseRand transferValueIncreaseGenerator
}

// transferValueIncreaseGenerator produces the post-transfer market value increase percentage.
type transferValueIncreaseGenerator interface {
	GeneratePercent() int
}

// randomTransferValueIncreaseGenerator provides a concurrency-safe random percentage generator.
type randomTransferValueIncreaseGenerator struct {
	mu  sync.Mutex
	rng *rand.Rand
}

// NewMarketListingService creates a market listing service with runtime randomness enabled.
func NewMarketListingService(
	listingRepo marketListingRepository,
	playerRepo marketPlayerRepository,
	txManager MarketTransactionManager,
) (*MarketListingService, error) {
	generator := &randomTransferValueIncreaseGenerator{
		rng: rand.New(rand.NewSource(rand.Int63())),
	}

	return newMarketListingService(listingRepo, playerRepo, txManager, generator)
}

// newMarketListingService keeps the public constructor small while letting tests
// inject a deterministic transfer value generator instead of runtime randomness.
func newMarketListingService(
	listingRepo marketListingRepository,
	playerRepo marketPlayerRepository,
	txManager MarketTransactionManager,
	generator transferValueIncreaseGenerator,
) (*MarketListingService, error) {
	if listingRepo == nil {
		return nil, ErrMarketListingRepositoryNil
	}

	if playerRepo == nil {
		return nil, ErrMarketPlayerRepositoryNil
	}

	if txManager == nil {
		return nil, ErrMarketTransactionManagerNil
	}

	if generator == nil {
		return nil, errors.New("transfer value increase generator is required")
	}

	return &MarketListingService{
		listingRepo:       listingRepo,
		playerRepo:        playerRepo,
		txManager:         txManager,
		valueIncreaseRand: generator,
	}, nil
}

// ListActive returns all active market listings.
func (s *MarketListingService) ListActive(ctx context.Context) ([]domain.MarketListing, error) {
	return s.listingRepo.ListActive(ctx)
}

// CreateListing validates ownership and creates a new active listing for the player.
func (s *MarketListingService) CreateListing(
	ctx context.Context,
	sellerTeamID, playerID, askingPrice int64,
) (*domain.MarketListing, error) {
	if askingPrice <= 0 {
		return nil, ErrInvalidAskingPrice
	}

	player, err := s.playerRepo.GetByID(ctx, playerID)
	if err != nil {
		return nil, err
	}

	if player.TeamID != sellerTeamID {
		return nil, ErrPlayerNotOwned
	}

	_, err = s.listingRepo.GetActiveByPlayerID(ctx, playerID)
	if err == nil {
		return nil, ErrListingAlreadyActive
	}
	if !errors.Is(err, repository.ErrMarketListingNotFound) {
		return nil, err
	}

	listing := &domain.MarketListing{
		PlayerID:     playerID,
		SellerTeamID: sellerTeamID,
		AskingPrice:  askingPrice,
		Status:       domain.MarketListingStatusActive,
	}

	if err := s.listingRepo.Create(ctx, listing); err != nil {
		return nil, err
	}

	return s.listingRepo.GetByID(ctx, listing.ID)
}

// CancelListing marks the player's active listing as cancelled.
func (s *MarketListingService) CancelListing(
	ctx context.Context,
	sellerTeamID, playerID int64,
) (*domain.MarketListing, error) {
	listing, err := s.listingRepo.GetActiveByPlayerID(ctx, playerID)
	if err != nil {
		return nil, err
	}

	player, err := s.playerRepo.GetByID(ctx, listing.PlayerID)
	if err != nil {
		return nil, err
	}

	if player.TeamID != sellerTeamID {
		return nil, ErrPlayerNotOwned
	}

	listing.Status = domain.MarketListingStatusCancelled
	if err := s.listingRepo.Update(ctx, listing); err != nil {
		return nil, err
	}

	return s.listingRepo.GetByID(ctx, listing.ID)
}

// BuyListing purchases an active listing inside a transaction and records the transfer history.
func (s *MarketListingService) BuyListing(
	ctx context.Context,
	buyerTeamID, listingID int64,
) (*domain.Transfer, error) {
	var transfer *domain.Transfer

	err := s.txManager.WithinTransaction(ctx, func(ctx context.Context, deps MarketTxDeps) error {
		listing, err := deps.ListingRepo.GetActiveByIDForUpdate(ctx, listingID)
		if err != nil {
			return err
		}

		player, err := deps.PlayerRepo.GetByIDForUpdate(ctx, listing.PlayerID)
		if err != nil {
			return err
		}

		if listing.SellerTeamID == buyerTeamID {
			return ErrCannotBuyOwnPlayer
		}

		buyerTeam, sellerTeam, err := lockTeamsForTransfer(ctx, deps.TeamRepo, buyerTeamID, listing.SellerTeamID)
		if err != nil {
			return err
		}

		if player.TeamID != sellerTeam.ID {
			return repository.ErrMarketListingNotFound
		}

		if buyerTeam.Budget < listing.AskingPrice {
			return ErrInsufficientBudget
		}

		oldMarketValue := player.MarketValue
		newMarketValue := increaseMarketValue(oldMarketValue, s.valueIncreaseRand.GeneratePercent())

		buyerTeam.Budget -= listing.AskingPrice
		sellerTeam.Budget += listing.AskingPrice
		if err := deps.TeamRepo.Update(ctx, buyerTeam); err != nil {
			return err
		}
		if err := deps.TeamRepo.Update(ctx, sellerTeam); err != nil {
			return err
		}

		player.TeamID = buyerTeam.ID
		player.MarketValue = newMarketValue
		if err := deps.PlayerRepo.Update(ctx, player); err != nil {
			return err
		}

		listing.Status = domain.MarketListingStatusSold
		if err := deps.ListingRepo.Update(ctx, listing); err != nil {
			return err
		}

		transfer = &domain.Transfer{
			PlayerID:       player.ID,
			SellerTeamID:   sellerTeam.ID,
			BuyerTeamID:    buyerTeam.ID,
			Price:          listing.AskingPrice,
			OldMarketValue: oldMarketValue,
			NewMarketValue: newMarketValue,
		}
		if err := deps.TransferRepo.Create(ctx, transfer); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return transfer, nil
}

// GeneratePercent returns a random inclusive percentage in the configured transfer growth range.
func (g *randomTransferValueIncreaseGenerator) GeneratePercent() int {
	g.mu.Lock()
	defer g.mu.Unlock()

	return domain.MinTransferValueIncreasePercent + g.rng.Intn(domain.MaxTransferValueIncreasePercent-domain.MinTransferValueIncreasePercent+1)
}

// lockTeamsForTransfer acquires team row locks in a stable order to avoid deadlocks.
func lockTeamsForTransfer(
	ctx context.Context,
	teamRepo marketTeamRepository,
	buyerTeamID, sellerTeamID int64,
) (*domain.Team, *domain.Team, error) {
	firstID, secondID := buyerTeamID, sellerTeamID
	if firstID > secondID {
		firstID, secondID = secondID, firstID
	}

	firstTeam, err := teamRepo.GetByIDForUpdate(ctx, firstID)
	if err != nil {
		return nil, nil, err
	}

	secondTeam, err := teamRepo.GetByIDForUpdate(ctx, secondID)
	if err != nil {
		return nil, nil, err
	}

	if buyerTeamID == firstID {
		return firstTeam, secondTeam, nil
	}

	return secondTeam, firstTeam, nil
}

// increaseMarketValue applies a percentage increase to the current market value.
func increaseMarketValue(currentValue int64, increasePercent int) int64 {
	return currentValue + currentValue*int64(increasePercent)/100
}
