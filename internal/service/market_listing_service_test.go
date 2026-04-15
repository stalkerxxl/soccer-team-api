package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stalkerxxl/soccer-team-api/internal/domain"
	"github.com/stalkerxxl/soccer-team-api/internal/repository"
)

func TestNewMarketListingServiceRequiresDependencies(t *testing.T) {
	t.Parallel()

	_, err := newMarketListingService(nil, &stubMarketPlayerRepository{}, &stubMarketTransactionManager{}, &stubTransferValueIncreaseGenerator{})
	if !errors.Is(err, ErrMarketListingRepositoryNil) {
		t.Fatalf("newMarketListingService() listing repo error = %v, want %v", err, ErrMarketListingRepositoryNil)
	}

	_, err = newMarketListingService(&stubMarketListingRepository{}, nil, &stubMarketTransactionManager{}, &stubTransferValueIncreaseGenerator{})
	if !errors.Is(err, ErrMarketPlayerRepositoryNil) {
		t.Fatalf("newMarketListingService() player repo error = %v, want %v", err, ErrMarketPlayerRepositoryNil)
	}

	_, err = newMarketListingService(&stubMarketListingRepository{}, &stubMarketPlayerRepository{}, nil, &stubTransferValueIncreaseGenerator{})
	if !errors.Is(err, ErrMarketTransactionManagerNil) {
		t.Fatalf("newMarketListingService() tx manager error = %v, want %v", err, ErrMarketTransactionManagerNil)
	}
}

func TestMarketListingServiceCreateListing(t *testing.T) {
	t.Parallel()

	listingRepo := &stubMarketListingRepository{
		getActiveByPlayerErr: repository.ErrMarketListingNotFound,
		listingByID: &domain.MarketListing{
			ID:           15,
			PlayerID:     7,
			SellerTeamID: 10,
			AskingPrice:  2500000,
			Status:       domain.MarketListingStatusActive,
		},
	}
	playerRepo := &stubMarketPlayerRepository{playerByID: &domain.Player{ID: 7, TeamID: 10}}
	svc := newTestMarketListingService(t, listingRepo, playerRepo, &stubMarketTransactionManager{}, &stubTransferValueIncreaseGenerator{})

	listing, err := svc.CreateListing(context.Background(), 10, 7, 2500000)
	if err != nil {
		t.Fatalf("CreateListing() error = %v", err)
	}

	if listing == nil {
		t.Fatal("CreateListing() returned nil listing")
	}

	if listingRepo.createdListing == nil {
		t.Fatal("CreateListing() did not call repository Create")
	}

	if listingRepo.createdListing.Status != domain.MarketListingStatusActive {
		t.Fatalf("created listing status = %q, want %q", listingRepo.createdListing.Status, domain.MarketListingStatusActive)
	}

	if listingRepo.createdListing.AskingPrice != 2500000 {
		t.Fatalf("created listing asking price = %d, want %d", listingRepo.createdListing.AskingPrice, 2500000)
	}
	if playerRepo.lastPlayerID != 7 {
		t.Fatalf("GetByID() player id = %d, want %d", playerRepo.lastPlayerID, 7)
	}
}

func TestMarketListingServiceCreateListingRejectsInvalidPrice(t *testing.T) {
	t.Parallel()

	svc := newTestMarketListingService(t, &stubMarketListingRepository{}, &stubMarketPlayerRepository{}, &stubMarketTransactionManager{}, &stubTransferValueIncreaseGenerator{})

	_, err := svc.CreateListing(context.Background(), 10, 7, 0)
	if !errors.Is(err, ErrInvalidAskingPrice) {
		t.Fatalf("CreateListing() error = %v, want %v", err, ErrInvalidAskingPrice)
	}
}

func TestMarketListingServiceCreateListingRejectsForeignPlayer(t *testing.T) {
	t.Parallel()

	svc := newTestMarketListingService(
		t,
		&stubMarketListingRepository{},
		&stubMarketPlayerRepository{playerByID: &domain.Player{ID: 7, TeamID: 99}},
		&stubMarketTransactionManager{},
		&stubTransferValueIncreaseGenerator{},
	)

	_, err := svc.CreateListing(context.Background(), 10, 7, 1000000)
	if !errors.Is(err, ErrPlayerNotOwned) {
		t.Fatalf("CreateListing() error = %v, want %v", err, ErrPlayerNotOwned)
	}
}

func TestMarketListingServiceCreateListingRejectsAlreadyActive(t *testing.T) {
	t.Parallel()

	svc := newTestMarketListingService(
		t,
		&stubMarketListingRepository{activeListingByPlayer: &domain.MarketListing{ID: 15, PlayerID: 7, Status: domain.MarketListingStatusActive}},
		&stubMarketPlayerRepository{playerByID: &domain.Player{ID: 7, TeamID: 10}},
		&stubMarketTransactionManager{},
		&stubTransferValueIncreaseGenerator{},
	)

	_, err := svc.CreateListing(context.Background(), 10, 7, 1000000)
	if !errors.Is(err, ErrListingAlreadyActive) {
		t.Fatalf("CreateListing() error = %v, want %v", err, ErrListingAlreadyActive)
	}
}

func TestMarketListingServiceCancelListing(t *testing.T) {
	t.Parallel()

	listingRepo := &stubMarketListingRepository{
		activeListingByPlayer: &domain.MarketListing{ID: 15, PlayerID: 7, SellerTeamID: 10, AskingPrice: 2500000, Status: domain.MarketListingStatusActive},
		listingByID:           &domain.MarketListing{ID: 15, PlayerID: 7, SellerTeamID: 10, AskingPrice: 2500000, Status: domain.MarketListingStatusCancelled},
	}
	playerRepo := &stubMarketPlayerRepository{playerByID: &domain.Player{ID: 7, TeamID: 10}}
	svc := newTestMarketListingService(t, listingRepo, playerRepo, &stubMarketTransactionManager{}, &stubTransferValueIncreaseGenerator{})

	listing, err := svc.CancelListing(context.Background(), 10, 7)
	if err != nil {
		t.Fatalf("CancelListing() error = %v", err)
	}

	if listing == nil {
		t.Fatal("CancelListing() returned nil listing")
	}

	if listingRepo.updatedListing == nil {
		t.Fatal("CancelListing() did not call repository Update")
	}

	if listingRepo.updatedListing.Status != domain.MarketListingStatusCancelled {
		t.Fatalf("updated listing status = %q, want %q", listingRepo.updatedListing.Status, domain.MarketListingStatusCancelled)
	}
}

func TestMarketListingServiceCancelListingRejectsForeignPlayer(t *testing.T) {
	t.Parallel()

	svc := newTestMarketListingService(
		t,
		&stubMarketListingRepository{activeListingByPlayer: &domain.MarketListing{ID: 15, PlayerID: 7, Status: domain.MarketListingStatusActive}},
		&stubMarketPlayerRepository{playerByID: &domain.Player{ID: 7, TeamID: 99}},
		&stubMarketTransactionManager{},
		&stubTransferValueIncreaseGenerator{},
	)

	_, err := svc.CancelListing(context.Background(), 10, 7)
	if !errors.Is(err, ErrPlayerNotOwned) {
		t.Fatalf("CancelListing() error = %v, want %v", err, ErrPlayerNotOwned)
	}
}

func TestMarketListingServiceListActive(t *testing.T) {
	t.Parallel()

	listingRepo := &stubMarketListingRepository{
		listings: []domain.MarketListing{{ID: 1, Status: domain.MarketListingStatusActive}, {ID: 2, Status: domain.MarketListingStatusActive}},
	}
	svc := newTestMarketListingService(t, listingRepo, &stubMarketPlayerRepository{}, &stubMarketTransactionManager{}, &stubTransferValueIncreaseGenerator{})

	listings, err := svc.ListActive(context.Background())
	if err != nil {
		t.Fatalf("ListActive() error = %v", err)
	}

	if len(listings) != 2 {
		t.Fatalf("len(listings) = %d, want %d", len(listings), 2)
	}
}

func TestMarketListingServiceBuyListing(t *testing.T) {
	t.Parallel()

	txListingRepo := &stubMarketListingRepository{
		activeListingByID: &domain.MarketListing{ID: 15, PlayerID: 7, SellerTeamID: 10, AskingPrice: 2500000, Status: domain.MarketListingStatusActive},
	}
	txPlayerRepo := &stubMarketPlayerRepository{
		playerByIDForUpdate: &domain.Player{ID: 7, TeamID: 10, MarketValue: 1000000},
	}
	txTeamRepo := &stubMarketTeamRepository{
		teamsByID: map[int64]*domain.Team{
			10: {ID: 10, Budget: 5000000},
			20: {ID: 20, Budget: 6000000},
		},
	}
	txTransferRepo := &stubMarketTransferRepository{}
	txManager := &stubMarketTransactionManager{deps: MarketTxDeps{
		ListingRepo:  txListingRepo,
		PlayerRepo:   txPlayerRepo,
		TeamRepo:     txTeamRepo,
		TransferRepo: txTransferRepo,
	}}
	svc := newTestMarketListingService(
		t,
		&stubMarketListingRepository{},
		&stubMarketPlayerRepository{},
		txManager,
		&stubTransferValueIncreaseGenerator{percent: 50},
	)

	transfer, err := svc.BuyListing(context.Background(), 20, 15)
	if err != nil {
		t.Fatalf("BuyListing() error = %v", err)
	}

	if transfer == nil {
		t.Fatal("BuyListing() returned nil transfer")
	}

	if txListingRepo.updatedListing == nil || txListingRepo.updatedListing.Status != domain.MarketListingStatusSold {
		t.Fatal("BuyListing() did not mark listing as sold")
	}

	if txPlayerRepo.updatedPlayer == nil {
		t.Fatal("BuyListing() did not update player")
	}

	if txPlayerRepo.updatedPlayer.TeamID != 20 {
		t.Fatalf("updated player team id = %d, want %d", txPlayerRepo.updatedPlayer.TeamID, 20)
	}

	if txPlayerRepo.updatedPlayer.MarketValue != 1500000 {
		t.Fatalf("updated player market value = %d, want %d", txPlayerRepo.updatedPlayer.MarketValue, 1500000)
	}

	buyer := txTeamRepo.teamsByID[20]
	seller := txTeamRepo.teamsByID[10]
	if buyer.Budget != 3500000 {
		t.Fatalf("buyer budget = %d, want %d", buyer.Budget, 3500000)
	}
	if seller.Budget != 7500000 {
		t.Fatalf("seller budget = %d, want %d", seller.Budget, 7500000)
	}

	if txTransferRepo.createdTransfer == nil {
		t.Fatal("BuyListing() did not create transfer")
	}

	if txTransferRepo.createdTransfer.Price != 2500000 {
		t.Fatalf("transfer price = %d, want %d", txTransferRepo.createdTransfer.Price, 2500000)
	}
}

func TestMarketListingServiceBuyListingRejectsOwnPlayer(t *testing.T) {
	t.Parallel()

	txManager := &stubMarketTransactionManager{deps: MarketTxDeps{
		ListingRepo:  &stubMarketListingRepository{activeListingByID: &domain.MarketListing{ID: 15, PlayerID: 7, SellerTeamID: 10, AskingPrice: 2500000, Status: domain.MarketListingStatusActive}},
		PlayerRepo:   &stubMarketPlayerRepository{playerByIDForUpdate: &domain.Player{ID: 7, TeamID: 10, MarketValue: 1000000}},
		TeamRepo:     &stubMarketTeamRepository{},
		TransferRepo: &stubMarketTransferRepository{},
	}}
	svc := newTestMarketListingService(t, &stubMarketListingRepository{}, &stubMarketPlayerRepository{}, txManager, &stubTransferValueIncreaseGenerator{percent: 50})

	_, err := svc.BuyListing(context.Background(), 10, 15)
	if !errors.Is(err, ErrCannotBuyOwnPlayer) {
		t.Fatalf("BuyListing() error = %v, want %v", err, ErrCannotBuyOwnPlayer)
	}
}

func TestMarketListingServiceBuyListingRejectsInsufficientBudget(t *testing.T) {
	t.Parallel()

	txManager := &stubMarketTransactionManager{deps: MarketTxDeps{
		ListingRepo: &stubMarketListingRepository{activeListingByID: &domain.MarketListing{ID: 15, PlayerID: 7, SellerTeamID: 10, AskingPrice: 2500000, Status: domain.MarketListingStatusActive}},
		PlayerRepo:  &stubMarketPlayerRepository{playerByIDForUpdate: &domain.Player{ID: 7, TeamID: 10, MarketValue: 1000000}},
		TeamRepo: &stubMarketTeamRepository{teamsByID: map[int64]*domain.Team{
			10: {ID: 10, Budget: 5000000},
			20: {ID: 20, Budget: 1000000},
		}},
		TransferRepo: &stubMarketTransferRepository{},
	}}
	svc := newTestMarketListingService(t, &stubMarketListingRepository{}, &stubMarketPlayerRepository{}, txManager, &stubTransferValueIncreaseGenerator{percent: 50})

	_, err := svc.BuyListing(context.Background(), 20, 15)
	if !errors.Is(err, ErrInsufficientBudget) {
		t.Fatalf("BuyListing() error = %v, want %v", err, ErrInsufficientBudget)
	}
}

func newTestMarketListingService(
	t *testing.T,
	listingRepo marketListingRepository,
	playerRepo marketPlayerRepository,
	txManager MarketTransactionManager,
	generator transferValueIncreaseGenerator,
) MarketListingManager {
	t.Helper()

	svc, err := newMarketListingService(listingRepo, playerRepo, txManager, generator)
	if err != nil {
		t.Fatalf("newMarketListingService() error = %v", err)
	}

	return svc
}

type stubMarketListingRepository struct {
	listingByID           *domain.MarketListing
	activeListingByID     *domain.MarketListing
	activeListingByPlayer *domain.MarketListing
	listings              []domain.MarketListing
	createdListing        *domain.MarketListing
	updatedListing        *domain.MarketListing
	getByIDErr            error
	getActiveByIDErr      error
	getActiveByPlayerErr  error
	listErr               error
	createErr             error
	updateErr             error
	nextID                int64
}

func (r *stubMarketListingRepository) Create(_ context.Context, listing *domain.MarketListing) error {
	r.createdListing = listing
	if r.nextID == 0 {
		r.nextID = 15
	}
	listing.ID = r.nextID
	return r.createErr
}

func (r *stubMarketListingRepository) GetByID(_ context.Context, _ int64) (*domain.MarketListing, error) {
	return r.listingByID, r.getByIDErr
}

func (r *stubMarketListingRepository) GetActiveByPlayerID(_ context.Context, _ int64) (*domain.MarketListing, error) {
	if r.getActiveByPlayerErr != nil {
		return nil, r.getActiveByPlayerErr
	}
	if r.activeListingByPlayer == nil {
		return nil, repository.ErrMarketListingNotFound
	}
	return r.activeListingByPlayer, nil
}

func (r *stubMarketListingRepository) GetActiveByIDForUpdate(_ context.Context, _ int64) (*domain.MarketListing, error) {
	if r.getActiveByIDErr != nil {
		return nil, r.getActiveByIDErr
	}
	if r.activeListingByID == nil {
		return nil, repository.ErrMarketListingNotFound
	}
	return r.activeListingByID, nil
}

func (r *stubMarketListingRepository) ListActive(_ context.Context) ([]domain.MarketListing, error) {
	return r.listings, r.listErr
}

func (r *stubMarketListingRepository) Update(_ context.Context, listing *domain.MarketListing) error {
	r.updatedListing = listing
	return r.updateErr
}

type stubMarketPlayerRepository struct {
	playerByID          *domain.Player
	playerByIDForUpdate *domain.Player
	updatedPlayer       *domain.Player
	lastPlayerID        int64
	getByIDErr          error
	getByIDForUpdateErr error
	updateErr           error
}

func (r *stubMarketPlayerRepository) GetByID(_ context.Context, id int64) (*domain.Player, error) {
	r.lastPlayerID = id
	return r.playerByID, r.getByIDErr
}

func (r *stubMarketPlayerRepository) GetByIDForUpdate(_ context.Context, id int64) (*domain.Player, error) {
	r.lastPlayerID = id
	return r.playerByIDForUpdate, r.getByIDForUpdateErr
}

func (r *stubMarketPlayerRepository) Update(_ context.Context, player *domain.Player) error {
	r.updatedPlayer = player
	return r.updateErr
}

type stubMarketTeamRepository struct {
	teamsByID map[int64]*domain.Team
	updateErr error
}

func (r *stubMarketTeamRepository) GetByIDForUpdate(_ context.Context, id int64) (*domain.Team, error) {
	team, ok := r.teamsByID[id]
	if !ok {
		return nil, repository.ErrTeamNotFound
	}
	return team, nil
}

func (r *stubMarketTeamRepository) Update(_ context.Context, _ *domain.Team) error {
	return r.updateErr
}

type stubMarketTransferRepository struct {
	createdTransfer *domain.Transfer
	createErr       error
}

func (r *stubMarketTransferRepository) Create(_ context.Context, transfer *domain.Transfer) error {
	r.createdTransfer = transfer
	transfer.ID = 1
	return r.createErr
}

type stubMarketTransactionManager struct {
	deps MarketTxDeps
	err  error
}

func (m *stubMarketTransactionManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context, deps MarketTxDeps) error) error {
	if m.err != nil {
		return m.err
	}
	return fn(ctx, m.deps)
}

type stubTransferValueIncreaseGenerator struct {
	percent int
}

func (g *stubTransferValueIncreaseGenerator) GeneratePercent() int {
	if g.percent == 0 {
		return domain.MinTransferValueIncreasePercent
	}
	return g.percent
}
