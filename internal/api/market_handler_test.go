package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stalkerxxl/soccer-team-api/internal/api/middleware"
	"github.com/stalkerxxl/soccer-team-api/internal/auth"
	"github.com/stalkerxxl/soccer-team-api/internal/domain"
	"github.com/stalkerxxl/soccer-team-api/internal/repository"
	"github.com/stalkerxxl/soccer-team-api/internal/service"
)

func TestMarketHandlerListPlayersIsPublic(t *testing.T) {
	t.Parallel()

	marketService := &stubAPIMarketService{
		listings: []domain.MarketListing{{
			ID:           15,
			PlayerID:     7,
			SellerTeamID: 10,
			AskingPrice:  2500000,
			Status:       domain.MarketListingStatusActive,
			Player: &domain.Player{
				ID:        7,
				FirstName: "Khvicha",
				LastName:  "Kvaratskhelia",
				Country:   &domain.Country{ID: 2, Name: "Georgia", ISOCode: "GE"},
			},
		}},
	}

	router := NewRouter(nil, nil, NewMarketHandler(&stubAPITeamService{}, marketService, apiTestLocalizer), middleware.Auth(&stubAPITokenProvider{}, apiTestLocalizer))

	req := httptest.NewRequest(http.MethodGet, "/market/players", nil)
	rrw := httptest.NewRecorder()

	router.ServeHTTP(rrw, req)

	if rrw.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rrw.Code, http.StatusOK)
	}

	var response []domain.MarketListing
	if err := json.Unmarshal(rrw.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if len(response) != 1 {
		t.Fatalf("len(response) = %d, want %d", len(response), 1)
	}
}

func TestMarketHandlerCreateListing(t *testing.T) {
	t.Parallel()

	teamService := &stubAPITeamService{teamByUserID: &domain.Team{ID: 10, UserID: 42}}
	marketService := &stubAPIMarketService{createdListing: &domain.MarketListing{
		ID:           15,
		PlayerID:     7,
		SellerTeamID: 10,
		AskingPrice:  2500000,
		Status:       domain.MarketListingStatusActive,
	}}
	tokenProvider := &stubAPITokenProvider{claims: auth.TokenClaims{UserID: 42}}

	router := NewRouter(nil, nil, NewMarketHandler(teamService, marketService, apiTestLocalizer), middleware.Auth(tokenProvider, apiTestLocalizer))

	body := bytes.NewBufferString(`{"asking_price":2500000}`)
	req := httptest.NewRequest(http.MethodPost, "/market/players/7/list", body)
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("Content-Type", "application/json")
	rrw := httptest.NewRecorder()

	router.ServeHTTP(rrw, req)

	if rrw.Code != http.StatusCreated {
		t.Fatalf("status code = %d, want %d", rrw.Code, http.StatusCreated)
	}

	if marketService.lastSellerTeamID != 10 {
		t.Fatalf("CreateListing() seller team id = %d, want %d", marketService.lastSellerTeamID, 10)
	}

	if marketService.lastPlayerID != 7 {
		t.Fatalf("CreateListing() player id = %d, want %d", marketService.lastPlayerID, 7)
	}

	if marketService.lastAskingPrice != 2500000 {
		t.Fatalf("CreateListing() asking price = %d, want %d", marketService.lastAskingPrice, 2500000)
	}
}

func TestMarketHandlerCancelListing(t *testing.T) {
	t.Parallel()

	teamService := &stubAPITeamService{teamByUserID: &domain.Team{ID: 10, UserID: 42}}
	marketService := &stubAPIMarketService{cancelledListing: &domain.MarketListing{
		ID:           15,
		PlayerID:     7,
		SellerTeamID: 10,
		AskingPrice:  2500000,
		Status:       domain.MarketListingStatusCancelled,
	}}
	tokenProvider := &stubAPITokenProvider{claims: auth.TokenClaims{UserID: 42}}

	router := NewRouter(nil, nil, NewMarketHandler(teamService, marketService, apiTestLocalizer), middleware.Auth(tokenProvider, apiTestLocalizer))

	req := httptest.NewRequest(http.MethodDelete, "/market/players/7/list", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rrw := httptest.NewRecorder()

	router.ServeHTTP(rrw, req)

	if rrw.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rrw.Code, http.StatusOK)
	}

	if marketService.lastSellerTeamID != 10 {
		t.Fatalf("CancelListing() seller team id = %d, want %d", marketService.lastSellerTeamID, 10)
	}

	if marketService.lastPlayerID != 7 {
		t.Fatalf("CancelListing() player id = %d, want %d", marketService.lastPlayerID, 7)
	}
}

func TestMarketHandlerBuyListing(t *testing.T) {
	t.Parallel()

	teamService := &stubAPITeamService{teamByUserID: &domain.Team{ID: 20, UserID: 42}}
	marketService := &stubAPIMarketService{boughtTransfer: &domain.Transfer{
		ID:             1,
		PlayerID:       7,
		SellerTeamID:   10,
		BuyerTeamID:    20,
		Price:          2500000,
		OldMarketValue: 1000000,
		NewMarketValue: 1500000,
	}}
	tokenProvider := &stubAPITokenProvider{claims: auth.TokenClaims{UserID: 42}}

	router := NewRouter(nil, nil, NewMarketHandler(teamService, marketService, apiTestLocalizer), middleware.Auth(tokenProvider, apiTestLocalizer))

	req := httptest.NewRequest(http.MethodPost, "/market/listings/15/buy", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rrw := httptest.NewRecorder()

	router.ServeHTTP(rrw, req)

	if rrw.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rrw.Code, http.StatusOK)
	}

	if marketService.lastBuyerTeamID != 20 {
		t.Fatalf("BuyListing() buyer team id = %d, want %d", marketService.lastBuyerTeamID, 20)
	}

	if marketService.lastListingID != 15 {
		t.Fatalf("BuyListing() listing id = %d, want %d", marketService.lastListingID, 15)
	}
}

func TestMarketHandlerCreateListingRequiresAccessToken(t *testing.T) {
	t.Parallel()

	router := NewRouter(nil, nil, NewMarketHandler(&stubAPITeamService{}, &stubAPIMarketService{}, apiTestLocalizer), middleware.Auth(&stubAPITokenProvider{}, apiTestLocalizer))

	body := bytes.NewBufferString(`{"asking_price":2500000}`)
	req := httptest.NewRequest(http.MethodPost, "/market/players/7/list", body)
	req.Header.Set("Content-Type", "application/json")
	rrw := httptest.NewRecorder()

	router.ServeHTTP(rrw, req)

	if rrw.Code != http.StatusUnauthorized {
		t.Fatalf("status code = %d, want %d", rrw.Code, http.StatusUnauthorized)
	}
}

func TestMarketHandlerBuyListingRequiresAccessToken(t *testing.T) {
	t.Parallel()

	router := NewRouter(nil, nil, NewMarketHandler(&stubAPITeamService{}, &stubAPIMarketService{}, apiTestLocalizer), middleware.Auth(&stubAPITokenProvider{}, apiTestLocalizer))

	req := httptest.NewRequest(http.MethodPost, "/market/listings/15/buy", nil)
	rrw := httptest.NewRecorder()

	router.ServeHTTP(rrw, req)

	if rrw.Code != http.StatusUnauthorized {
		t.Fatalf("status code = %d, want %d", rrw.Code, http.StatusUnauthorized)
	}
}

func TestMarketHandlerReturnsBadRequestForInvalidPrice(t *testing.T) {
	t.Parallel()

	teamService := &stubAPITeamService{teamByUserID: &domain.Team{ID: 10, UserID: 42}}
	marketService := &stubAPIMarketService{createErr: service.ErrInvalidAskingPrice}
	tokenProvider := &stubAPITokenProvider{claims: auth.TokenClaims{UserID: 42}}

	router := NewRouter(nil, nil, NewMarketHandler(teamService, marketService, apiTestLocalizer), middleware.Auth(tokenProvider, apiTestLocalizer))

	body := bytes.NewBufferString(`{"asking_price":0}`)
	req := httptest.NewRequest(http.MethodPost, "/market/players/7/list", body)
	req.Header.Set("Accept-Language", "ka-GE,ka;q=0.9,en;q=0.8")
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("Content-Type", "application/json")
	rrw := httptest.NewRecorder()

	router.ServeHTTP(rrw, req)

	if rrw.Code != http.StatusBadRequest {
		t.Fatalf("status code = %d, want %d", rrw.Code, http.StatusBadRequest)
	}

	var response errorResponse
	if err := json.Unmarshal(rrw.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if response.Code != "invalid_asking_price" {
		t.Fatalf("response.Code = %q, want %q", response.Code, "invalid_asking_price")
	}

	if response.Message != "მოთხოვნილი ფასი ნულზე მეტი უნდა იყოს." {
		t.Fatalf("response.Message = %q, want %q", response.Message, "მოთხოვნილი ფასი ნულზე მეტი უნდა იყოს.")
	}
}

func TestMarketHandlerReturnsForbiddenForForeignPlayer(t *testing.T) {
	t.Parallel()

	teamService := &stubAPITeamService{teamByUserID: &domain.Team{ID: 10, UserID: 42}}
	marketService := &stubAPIMarketService{createErr: service.ErrPlayerNotOwned}
	tokenProvider := &stubAPITokenProvider{claims: auth.TokenClaims{UserID: 42}}

	router := NewRouter(nil, nil, NewMarketHandler(teamService, marketService, apiTestLocalizer), middleware.Auth(tokenProvider, apiTestLocalizer))

	body := bytes.NewBufferString(`{"asking_price":2500000}`)
	req := httptest.NewRequest(http.MethodPost, "/market/players/7/list", body)
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("Content-Type", "application/json")
	rrw := httptest.NewRecorder()

	router.ServeHTTP(rrw, req)

	if rrw.Code != http.StatusForbidden {
		t.Fatalf("status code = %d, want %d", rrw.Code, http.StatusForbidden)
	}
}

func TestMarketHandlerReturnsListingNotFoundOnCancel(t *testing.T) {
	t.Parallel()

	teamService := &stubAPITeamService{teamByUserID: &domain.Team{ID: 10, UserID: 42}}
	marketService := &stubAPIMarketService{cancelErr: repository.ErrMarketListingNotFound}
	tokenProvider := &stubAPITokenProvider{claims: auth.TokenClaims{UserID: 42}}

	router := NewRouter(nil, nil, NewMarketHandler(teamService, marketService, apiTestLocalizer), middleware.Auth(tokenProvider, apiTestLocalizer))

	req := httptest.NewRequest(http.MethodDelete, "/market/players/7/list", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rrw := httptest.NewRecorder()

	router.ServeHTTP(rrw, req)

	if rrw.Code != http.StatusNotFound {
		t.Fatalf("status code = %d, want %d", rrw.Code, http.StatusNotFound)
	}

	var response errorResponse
	if err := json.Unmarshal(rrw.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if response.Code != "listing_not_found" {
		t.Fatalf("response.Code = %q, want %q", response.Code, "listing_not_found")
	}
}

func TestMarketHandlerReturnsConflictForInsufficientBudget(t *testing.T) {
	t.Parallel()

	teamService := &stubAPITeamService{teamByUserID: &domain.Team{ID: 20, UserID: 42}}
	marketService := &stubAPIMarketService{buyErr: service.ErrInsufficientBudget}
	tokenProvider := &stubAPITokenProvider{claims: auth.TokenClaims{UserID: 42}}

	router := NewRouter(nil, nil, NewMarketHandler(teamService, marketService, apiTestLocalizer), middleware.Auth(tokenProvider, apiTestLocalizer))

	req := httptest.NewRequest(http.MethodPost, "/market/listings/15/buy", nil)
	req.Header.Set("Accept-Language", "fr")
	req.Header.Set("Authorization", "Bearer valid-token")
	rrw := httptest.NewRecorder()

	router.ServeHTTP(rrw, req)

	if rrw.Code != http.StatusConflict {
		t.Fatalf("status code = %d, want %d", rrw.Code, http.StatusConflict)
	}

	var response errorResponse
	if err := json.Unmarshal(rrw.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if response.Code != "insufficient_budget" {
		t.Fatalf("response.Code = %q, want %q", response.Code, "insufficient_budget")
	}

	if response.Message != "Insufficient budget." {
		t.Fatalf("response.Message = %q, want %q", response.Message, "Insufficient budget.")
	}
}

func TestMarketHandlerReturnsForbiddenForOwnPlayerBuy(t *testing.T) {
	t.Parallel()

	teamService := &stubAPITeamService{teamByUserID: &domain.Team{ID: 10, UserID: 42}}
	marketService := &stubAPIMarketService{buyErr: service.ErrCannotBuyOwnPlayer}
	tokenProvider := &stubAPITokenProvider{claims: auth.TokenClaims{UserID: 42}}

	router := NewRouter(nil, nil, NewMarketHandler(teamService, marketService, apiTestLocalizer), middleware.Auth(tokenProvider, apiTestLocalizer))

	req := httptest.NewRequest(http.MethodPost, "/market/listings/15/buy", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rrw := httptest.NewRecorder()

	router.ServeHTTP(rrw, req)

	if rrw.Code != http.StatusForbidden {
		t.Fatalf("status code = %d, want %d", rrw.Code, http.StatusForbidden)
	}
}

type stubAPIMarketService struct {
	listings         []domain.MarketListing
	createdListing   *domain.MarketListing
	cancelledListing *domain.MarketListing
	boughtTransfer   *domain.Transfer
	lastSellerTeamID int64
	lastBuyerTeamID  int64
	lastPlayerID     int64
	lastListingID    int64
	lastAskingPrice  int64
	listErr          error
	createErr        error
	cancelErr        error
	buyErr           error
}

func (s *stubAPIMarketService) ListActive(_ context.Context) ([]domain.MarketListing, error) {
	return s.listings, s.listErr
}

func (s *stubAPIMarketService) CreateListing(_ context.Context, sellerTeamID, playerID, askingPrice int64) (*domain.MarketListing, error) {
	s.lastSellerTeamID = sellerTeamID
	s.lastPlayerID = playerID
	s.lastAskingPrice = askingPrice
	if s.createErr != nil {
		return nil, s.createErr
	}
	return s.createdListing, nil
}

func (s *stubAPIMarketService) CancelListing(_ context.Context, sellerTeamID, playerID int64) (*domain.MarketListing, error) {
	s.lastSellerTeamID = sellerTeamID
	s.lastPlayerID = playerID
	if s.cancelErr != nil {
		return nil, s.cancelErr
	}
	return s.cancelledListing, nil
}

func (s *stubAPIMarketService) BuyListing(_ context.Context, buyerTeamID, listingID int64) (*domain.Transfer, error) {
	s.lastBuyerTeamID = buyerTeamID
	s.lastListingID = listingID
	if s.buyErr != nil {
		return nil, s.buyErr
	}
	return s.boughtTransfer, nil
}
