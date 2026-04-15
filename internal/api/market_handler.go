package api

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/stalkerxxl/soccer-team-api/internal/i18n"
	"github.com/stalkerxxl/soccer-team-api/internal/repository"
	"github.com/stalkerxxl/soccer-team-api/internal/service"
)

type MarketHandler struct {
	teamService   service.TeamManager
	marketService service.MarketListingManager
	localizer     *i18n.Localizer
}

type createMarketListingRequest struct {
	AskingPrice int64 `json:"asking_price"`
}

// NewMarketHandler creates a market HTTP handler with localized error responses.
func NewMarketHandler(teamService service.TeamManager, marketService service.MarketListingManager, localizer *i18n.Localizer) *MarketHandler {
	return &MarketHandler{
		teamService:   teamService,
		marketService: marketService,
		localizer:     localizer,
	}
}

// ListPlayers returns the public list of active market listings.
func (h *MarketHandler) ListPlayers(w http.ResponseWriter, r *http.Request) {
	listings, err := h.marketService.ListActive(r.Context())
	if err != nil {
		h.writeMarketError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, listings)
}

// CreateListing creates a new market listing for a player owned by the current team.
func (h *MarketHandler) CreateListing(w http.ResponseWriter, r *http.Request) {
	var req createMarketListingRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, h.localizer, http.StatusBadRequest, "invalid_request", "error.invalid_request")
		return
	}

	team, err := currentTeam(r, h.teamService)
	if err != nil {
		h.writeMarketError(w, r, err)
		return
	}

	playerID, err := playerIDFromRequest(r)
	if err != nil {
		h.writeMarketError(w, r, err)
		return
	}

	listing, err := h.marketService.CreateListing(r.Context(), team.ID, playerID, req.AskingPrice)
	if err != nil {
		h.writeMarketError(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, listing)
}

// CancelListing cancels the active market listing for a player owned by the current team.
func (h *MarketHandler) CancelListing(w http.ResponseWriter, r *http.Request) {
	team, err := currentTeam(r, h.teamService)
	if err != nil {
		h.writeMarketError(w, r, err)
		return
	}

	playerID, err := playerIDFromRequest(r)
	if err != nil {
		h.writeMarketError(w, r, err)
		return
	}

	listing, err := h.marketService.CancelListing(r.Context(), team.ID, playerID)
	if err != nil {
		h.writeMarketError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, listing)
}

// BuyListing buys an active listing on behalf of the current team.
func (h *MarketHandler) BuyListing(w http.ResponseWriter, r *http.Request) {
	team, err := currentTeam(r, h.teamService)
	if err != nil {
		h.writeMarketError(w, r, err)
		return
	}

	listingID, err := listingIDFromRequest(r)
	if err != nil {
		h.writeMarketError(w, r, err)
		return
	}

	transfer, err := h.marketService.BuyListing(r.Context(), team.ID, listingID)
	if err != nil {
		h.writeMarketError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, transfer)
}

// writeMarketError maps market-related errors to API responses.
func (h *MarketHandler) writeMarketError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidAccessToken):
		writeError(w, r, h.localizer, http.StatusUnauthorized, "invalid_access_token", "error.invalid_access_token")
	case errors.Is(err, errInvalidPlayerID):
		writeError(w, r, h.localizer, http.StatusBadRequest, "invalid_player_id", "error.invalid_player_id")
	case errors.Is(err, errInvalidListingID):
		writeError(w, r, h.localizer, http.StatusBadRequest, "invalid_listing_id", "error.invalid_listing_id")
	case errors.Is(err, service.ErrInvalidAskingPrice):
		writeError(w, r, h.localizer, http.StatusBadRequest, "invalid_asking_price", "error.invalid_asking_price")
	case errors.Is(err, service.ErrPlayerNotOwned):
		writeError(w, r, h.localizer, http.StatusForbidden, "player_not_owned", "error.player_not_owned")
	case errors.Is(err, service.ErrCannotBuyOwnPlayer):
		writeError(w, r, h.localizer, http.StatusForbidden, "cannot_buy_own_player", "error.cannot_buy_own_player")
	case errors.Is(err, service.ErrListingAlreadyActive):
		writeError(w, r, h.localizer, http.StatusConflict, "listing_already_active", "error.listing_already_active")
	case errors.Is(err, service.ErrInsufficientBudget):
		writeError(w, r, h.localizer, http.StatusConflict, "insufficient_budget", "error.insufficient_budget")
	case errors.Is(err, repository.ErrTeamNotFound):
		writeError(w, r, h.localizer, http.StatusNotFound, "team_not_found", "error.team_not_found")
	case errors.Is(err, repository.ErrPlayerNotFound):
		writeError(w, r, h.localizer, http.StatusNotFound, "player_not_found", "error.player_not_found")
	case errors.Is(err, repository.ErrMarketListingNotFound):
		writeError(w, r, h.localizer, http.StatusNotFound, "listing_not_found", "error.listing_not_found")
	default:
		writeError(w, r, h.localizer, http.StatusInternalServerError, "internal_error", "error.internal_error")
	}
}

var errInvalidListingID = errors.New("invalid listing id")

// listingIDFromRequest parses and validates a listing ID from the route params.
func listingIDFromRequest(r *http.Request) (int64, error) {
	listingIDParam := strings.TrimSpace(chi.URLParam(r, "listingId"))
	if listingIDParam == "" {
		return 0, errInvalidListingID
	}

	listingID, err := strconv.ParseInt(listingIDParam, 10, 64)
	if err != nil || listingID < 1 {
		return 0, errInvalidListingID
	}

	return listingID, nil
}
