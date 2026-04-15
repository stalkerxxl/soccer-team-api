package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stalkerxxl/soccer-team-api/internal/api/middleware"
	"github.com/stalkerxxl/soccer-team-api/internal/auth"
	"github.com/stalkerxxl/soccer-team-api/internal/domain"
	"github.com/stalkerxxl/soccer-team-api/internal/repository"
)

func TestMeHandlerGetTeam(t *testing.T) {
	t.Parallel()

	teamService := &stubAPITeamService{
		teamByUserID: &domain.Team{
			ID:                      10,
			UserID:                  42,
			Name:                    "Your team",
			CountryID:               1,
			Budget:                  domain.InitialTeamBudget,
			TotalPlayersMarketValue: 20_000_000,
			Country:                 &domain.Country{ID: 1, Name: "United Kingdom", ISOCode: "GB"},
		},
	}
	playerService := &stubAPIPlayerService{}
	tokenProvider := &stubAPITokenProvider{claims: auth.TokenClaims{UserID: 42}}

	router := NewRouter(nil, NewMeHandler(teamService, playerService, apiTestLocalizer), nil, middleware.Auth(tokenProvider, apiTestLocalizer))

	req := httptest.NewRequest(http.MethodGet, "/me/team", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rrw := httptest.NewRecorder()

	router.ServeHTTP(rrw, req)

	if rrw.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rrw.Code, http.StatusOK)
	}

	if teamService.lastUserID != 42 {
		t.Fatalf("GetByUserID() user id = %d, want %d", teamService.lastUserID, 42)
	}

	var response domain.Team
	if err := json.Unmarshal(rrw.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if response.ID != 10 {
		t.Fatalf("response.ID = %d, want %d", response.ID, 10)
	}

	if response.TotalPlayersMarketValue != 20_000_000 {
		t.Fatalf("response.TotalPlayersMarketValue = %d, want %d", response.TotalPlayersMarketValue, int64(20_000_000))
	}
}

func TestMeHandlerUpdateTeam(t *testing.T) {
	t.Parallel()

	teamService := &stubAPITeamService{
		teamByUserID: &domain.Team{ID: 10, UserID: 42, Name: "Your team", CountryID: 1, Budget: domain.InitialTeamBudget},
		teamByID:     &domain.Team{ID: 10, UserID: 42, Name: "Dinamo", CountryID: 2, Budget: domain.InitialTeamBudget},
	}
	tokenProvider := &stubAPITokenProvider{claims: auth.TokenClaims{UserID: 42}}

	router := NewRouter(nil, NewMeHandler(teamService, &stubAPIPlayerService{}, apiTestLocalizer), nil, middleware.Auth(tokenProvider, apiTestLocalizer))

	body := bytes.NewBufferString(`{"name":"Dinamo","country_id":2}`)
	req := httptest.NewRequest(http.MethodPatch, "/me/team", body)
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("Content-Type", "application/json")
	rrw := httptest.NewRecorder()

	router.ServeHTTP(rrw, req)

	if rrw.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rrw.Code, http.StatusOK)
	}

	if teamService.updatedTeam == nil {
		t.Fatal("Update() was not called")
	}

	if teamService.updatedTeam.Name != "Dinamo" {
		t.Fatalf("updated team name = %q, want %q", teamService.updatedTeam.Name, "Dinamo")
	}

	if teamService.updatedTeam.CountryID != 2 {
		t.Fatalf("updated team country id = %d, want %d", teamService.updatedTeam.CountryID, 2)
	}
}

func TestMeHandlerListPlayers(t *testing.T) {
	t.Parallel()

	teamService := &stubAPITeamService{
		teamByUserID: &domain.Team{ID: 10, UserID: 42},
	}
	playerService := &stubAPIPlayerService{
		playersByTeam: []domain.Player{
			{ID: 1, TeamID: 10, FirstName: "Liam", LastName: "Smith", Position: domain.PlayerPositionGoalkeeper},
			{ID: 2, TeamID: 10, FirstName: "Noah", LastName: "Brown", Position: domain.PlayerPositionDefender},
		},
	}
	tokenProvider := &stubAPITokenProvider{claims: auth.TokenClaims{UserID: 42}}

	router := NewRouter(nil, NewMeHandler(teamService, playerService, apiTestLocalizer), nil, middleware.Auth(tokenProvider, apiTestLocalizer))

	req := httptest.NewRequest(http.MethodGet, "/me/players", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rrw := httptest.NewRecorder()

	router.ServeHTTP(rrw, req)

	if rrw.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rrw.Code, http.StatusOK)
	}

	if playerService.lastTeamID != 10 {
		t.Fatalf("ListByTeamID() team id = %d, want %d", playerService.lastTeamID, 10)
	}

	var response []domain.Player
	if err := json.Unmarshal(rrw.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if len(response) != 2 {
		t.Fatalf("len(response) = %d, want %d", len(response), 2)
	}
}

func TestMeHandlerGetPlayer(t *testing.T) {
	t.Parallel()

	teamService := &stubAPITeamService{teamByUserID: &domain.Team{ID: 10, UserID: 42}}
	playerService := &stubAPIPlayerService{
		playerByID: &domain.Player{ID: 7, TeamID: 10, FirstName: "Nico", LastName: "Williams", CountryID: 2},
	}
	tokenProvider := &stubAPITokenProvider{claims: auth.TokenClaims{UserID: 42}}

	router := NewRouter(nil, NewMeHandler(teamService, playerService, apiTestLocalizer), nil, middleware.Auth(tokenProvider, apiTestLocalizer))

	req := httptest.NewRequest(http.MethodGet, "/me/players/7", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rrw := httptest.NewRecorder()

	router.ServeHTTP(rrw, req)

	if rrw.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rrw.Code, http.StatusOK)
	}

	if playerService.lastPlayerID != 7 {
		t.Fatalf("GetByID() player id = %d, want %d", playerService.lastPlayerID, 7)
	}
}

func TestMeHandlerUpdatePlayer(t *testing.T) {
	t.Parallel()

	teamService := &stubAPITeamService{teamByUserID: &domain.Team{ID: 10, UserID: 42}}
	playerService := &stubAPIPlayerService{
		playerByID: &domain.Player{ID: 7, TeamID: 10, FirstName: "Nico", LastName: "Williams", CountryID: 2},
		playerByIDSequence: []*domain.Player{
			{ID: 7, TeamID: 10, FirstName: "Nico", LastName: "Williams", CountryID: 2},
			{ID: 7, TeamID: 10, FirstName: "Khvicha", LastName: "Kvaratskhelia", CountryID: 3},
		},
	}
	tokenProvider := &stubAPITokenProvider{claims: auth.TokenClaims{UserID: 42}}

	router := NewRouter(nil, NewMeHandler(teamService, playerService, apiTestLocalizer), nil, middleware.Auth(tokenProvider, apiTestLocalizer))

	body := bytes.NewBufferString(`{"first_name":"Khvicha","last_name":"Kvaratskhelia","country_id":3}`)
	req := httptest.NewRequest(http.MethodPatch, "/me/players/7", body)
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("Content-Type", "application/json")
	rrw := httptest.NewRecorder()

	router.ServeHTTP(rrw, req)

	if rrw.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rrw.Code, http.StatusOK)
	}

	if playerService.updatedPlayer == nil {
		t.Fatal("Update() was not called")
	}

	if playerService.updatedPlayer.FirstName != "Khvicha" {
		t.Fatalf("updated player first name = %q, want %q", playerService.updatedPlayer.FirstName, "Khvicha")
	}
}

func TestMeHandlerReturnsPlayerNotFoundForForeignPlayer(t *testing.T) {
	t.Parallel()

	teamService := &stubAPITeamService{teamByUserID: &domain.Team{ID: 10, UserID: 42}}
	playerService := &stubAPIPlayerService{
		playerByID: &domain.Player{ID: 7, TeamID: 99, FirstName: "Nico", LastName: "Williams"},
	}
	tokenProvider := &stubAPITokenProvider{claims: auth.TokenClaims{UserID: 42}}

	router := NewRouter(nil, NewMeHandler(teamService, playerService, apiTestLocalizer), nil, middleware.Auth(tokenProvider, apiTestLocalizer))

	req := httptest.NewRequest(http.MethodGet, "/me/players/7", nil)
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

	if response.Code != "player_not_found" {
		t.Fatalf("response.Code = %q, want %q", response.Code, "player_not_found")
	}
}

func TestMeHandlerReturnsBadRequestForInvalidPlayerID(t *testing.T) {
	t.Parallel()

	teamService := &stubAPITeamService{teamByUserID: &domain.Team{ID: 10, UserID: 42}}
	tokenProvider := &stubAPITokenProvider{claims: auth.TokenClaims{UserID: 42}}

	router := NewRouter(nil, NewMeHandler(teamService, &stubAPIPlayerService{}, apiTestLocalizer), nil, middleware.Auth(tokenProvider, apiTestLocalizer))

	req := httptest.NewRequest(http.MethodGet, "/me/players/not-a-number", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rrw := httptest.NewRecorder()

	router.ServeHTTP(rrw, req)

	if rrw.Code != http.StatusBadRequest {
		t.Fatalf("status code = %d, want %d", rrw.Code, http.StatusBadRequest)
	}
}

func TestMeHandlerRequiresAccessToken(t *testing.T) {
	t.Parallel()

	router := NewRouter(nil, NewMeHandler(&stubAPITeamService{}, &stubAPIPlayerService{}, apiTestLocalizer), nil, middleware.Auth(&stubAPITokenProvider{}, apiTestLocalizer))

	req := httptest.NewRequest(http.MethodGet, "/me/team", nil)
	rrw := httptest.NewRecorder()

	router.ServeHTTP(rrw, req)

	if rrw.Code != http.StatusUnauthorized {
		t.Fatalf("status code = %d, want %d", rrw.Code, http.StatusUnauthorized)
	}
}

func TestMeHandlerLocalizesInvalidRequestToGeorgian(t *testing.T) {
	t.Parallel()

	teamService := &stubAPITeamService{teamByUserID: &domain.Team{ID: 10, UserID: 42}}
	tokenProvider := &stubAPITokenProvider{claims: auth.TokenClaims{UserID: 42}}

	router := NewRouter(nil, NewMeHandler(teamService, &stubAPIPlayerService{}, apiTestLocalizer), nil, middleware.Auth(tokenProvider, apiTestLocalizer))

	req := httptest.NewRequest(http.MethodPatch, "/me/team", bytes.NewBufferString(`{"unknown":"field"}`))
	req.Header.Set("Accept-Language", "ka")
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

	if response.Code != "invalid_request" {
		t.Fatalf("response.Code = %q, want %q", response.Code, "invalid_request")
	}

	if response.Message != "მოთხოვნის სხეული არასწორია." {
		t.Fatalf("response.Message = %q, want %q", response.Message, "მოთხოვნის სხეული არასწორია.")
	}
}

func TestMeHandlerReturnsNotFoundWhenTeamMissing(t *testing.T) {
	t.Parallel()

	teamService := &stubAPITeamService{getByUserIDErr: repository.ErrTeamNotFound}
	tokenProvider := &stubAPITokenProvider{claims: auth.TokenClaims{UserID: 42}}

	router := NewRouter(nil, NewMeHandler(teamService, &stubAPIPlayerService{}, apiTestLocalizer), nil, middleware.Auth(tokenProvider, apiTestLocalizer))

	req := httptest.NewRequest(http.MethodGet, "/me/team", nil)
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

	if response.Code != "team_not_found" {
		t.Fatalf("response.Code = %q, want %q", response.Code, "team_not_found")
	}
}

type stubAPITeamService struct {
	teamByID       *domain.Team
	teamByUserID   *domain.Team
	teams          []domain.Team
	countryByID    *domain.Country
	countries      []domain.Country
	updatedTeam    *domain.Team
	lastUserID     int64
	getByIDErr     error
	getAllErr      error
	getByUserIDErr error
	updateErr      error
	getCountryErr  error
	listErr        error
}

func (s *stubAPITeamService) CreateDefault(_ context.Context, _ int64) (*domain.Team, error) {
	return nil, nil
}

func (s *stubAPITeamService) GetByID(_ context.Context, _ int64) (*domain.Team, error) {
	return s.teamByID, s.getByIDErr
}

func (s *stubAPITeamService) GetAll(_ context.Context) ([]domain.Team, error) {
	return s.teams, s.getAllErr
}

func (s *stubAPITeamService) GetByUserID(_ context.Context, userID int64) (*domain.Team, error) {
	s.lastUserID = userID
	return s.teamByUserID, s.getByUserIDErr
}

func (s *stubAPITeamService) Update(_ context.Context, team *domain.Team) error {
	s.updatedTeam = team
	return s.updateErr
}

func (s *stubAPITeamService) GetCountryByID(_ context.Context, _ int64) (*domain.Country, error) {
	return s.countryByID, s.getCountryErr
}

func (s *stubAPITeamService) ListCountries(_ context.Context) ([]domain.Country, error) {
	return s.countries, s.listErr
}

type stubAPIPlayerService struct {
	playerByID         *domain.Player
	playerByIDSequence []*domain.Player
	playersByTeam      []domain.Player
	updatedPlayer      *domain.Player
	lastTeamID         int64
	lastPlayerID       int64
	getByIDErr         error
	listErr            error
	updateErr          error
	getByIDCallCount   int
}

func (s *stubAPIPlayerService) CreateInitialSquad(_ context.Context, _ int64) ([]domain.Player, error) {
	return nil, nil
}

func (s *stubAPIPlayerService) GetByID(_ context.Context, id int64) (*domain.Player, error) {
	s.lastPlayerID = id
	if s.getByIDErr != nil {
		return nil, s.getByIDErr
	}
	if len(s.playerByIDSequence) > 0 {
		index := s.getByIDCallCount
		if index >= len(s.playerByIDSequence) {
			index = len(s.playerByIDSequence) - 1
		}
		s.getByIDCallCount++
		return s.playerByIDSequence[index], nil
	}
	return s.playerByID, nil
}

func (s *stubAPIPlayerService) ListByTeamID(_ context.Context, teamID int64) ([]domain.Player, error) {
	s.lastTeamID = teamID
	return s.playersByTeam, s.listErr
}

func (s *stubAPIPlayerService) Update(_ context.Context, player *domain.Player) error {
	s.updatedPlayer = player
	return s.updateErr
}

type stubAPITokenProvider struct {
	claims   auth.TokenClaims
	parseErr error
}

func (p *stubAPITokenProvider) GenerateAccessToken(_ int64) (string, time.Time, error) {
	return "", time.Time{}, nil
}

func (p *stubAPITokenProvider) ParseAccessToken(_ string) (auth.TokenClaims, error) {
	return p.claims, p.parseErr
}
