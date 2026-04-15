package api

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/stalkerxxl/soccer-team-api/internal/domain"
	"github.com/stalkerxxl/soccer-team-api/internal/i18n"
	"github.com/stalkerxxl/soccer-team-api/internal/repository"
	"github.com/stalkerxxl/soccer-team-api/internal/service"
)

type MeHandler struct {
	teamService   service.TeamManager
	playerService service.PlayerManager
	localizer     *i18n.Localizer
}

type updateTeamRequest struct {
	Name      string `json:"name"`
	CountryID int64  `json:"country_id"`
}

type updatePlayerRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	CountryID int64  `json:"country_id"`
}

// NewMeHandler creates a handler for authenticated team and player endpoints.
func NewMeHandler(teamService service.TeamManager, playerService service.PlayerManager, localizer *i18n.Localizer) *MeHandler {
	return &MeHandler{
		teamService:   teamService,
		playerService: playerService,
		localizer:     localizer,
	}
}

// GetTeam returns the current user's team.
func (h *MeHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	team, err := h.currentTeam(r)
	if err != nil {
		h.writeMeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, team)
}

// UpdateTeam updates the current user's team fields and returns the fresh state.
func (h *MeHandler) UpdateTeam(w http.ResponseWriter, r *http.Request) {
	var req updateTeamRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, h.localizer, http.StatusBadRequest, "invalid_request", "error.invalid_request")
		return
	}

	team, err := h.currentTeam(r)
	if err != nil {
		h.writeMeError(w, r, err)
		return
	}

	team.Name = req.Name
	team.CountryID = req.CountryID

	if err := h.teamService.Update(r.Context(), team); err != nil {
		h.writeMeError(w, r, err)
		return
	}

	updatedTeam, err := h.teamService.GetByID(r.Context(), team.ID)
	if err != nil {
		h.writeMeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, updatedTeam)
}

// ListPlayers returns all players of the current user's team.
func (h *MeHandler) ListPlayers(w http.ResponseWriter, r *http.Request) {
	team, err := h.currentTeam(r)
	if err != nil {
		h.writeMeError(w, r, err)
		return
	}

	players, err := h.playerService.ListByTeamID(r.Context(), team.ID)
	if err != nil {
		h.writeMeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, players)
}

// GetPlayer returns a single player that belongs to the current user's team.
func (h *MeHandler) GetPlayer(w http.ResponseWriter, r *http.Request) {
	player, err := h.currentPlayer(r)
	if err != nil {
		h.writeMeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, player)
}

// UpdatePlayer updates a player that belongs to the current user's team.
func (h *MeHandler) UpdatePlayer(w http.ResponseWriter, r *http.Request) {
	var req updatePlayerRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, h.localizer, http.StatusBadRequest, "invalid_request", "error.invalid_request")
		return
	}

	player, err := h.currentPlayer(r)
	if err != nil {
		h.writeMeError(w, r, err)
		return
	}

	player.FirstName = req.FirstName
	player.LastName = req.LastName
	player.CountryID = req.CountryID

	if err := h.playerService.Update(r.Context(), player); err != nil {
		h.writeMeError(w, r, err)
		return
	}

	updatedPlayer, err := h.playerService.GetByID(r.Context(), player.ID)
	if err != nil {
		h.writeMeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, updatedPlayer)
}

// currentTeam resolves the current user's team from the authenticated request.
func (h *MeHandler) currentTeam(r *http.Request) (*domain.Team, error) {
	return currentTeam(r, h.teamService)
}

// currentPlayer resolves a player from the route and ensures it belongs to the current team.
func (h *MeHandler) currentPlayer(r *http.Request) (*domain.Player, error) {
	team, err := h.currentTeam(r)
	if err != nil {
		return nil, err
	}

	playerID, err := playerIDFromRequest(r)
	if err != nil {
		return nil, err
	}

	player, err := h.playerService.GetByID(r.Context(), playerID)
	if err != nil {
		return nil, err
	}

	// Treat foreign players as not found so the endpoint never exposes cross-team access.
	if player.TeamID != team.ID {
		return nil, repository.ErrPlayerNotFound
	}

	return player, nil
}

// writeMeError maps team and player endpoint errors to API responses.
func (h *MeHandler) writeMeError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidAccessToken):
		writeError(w, r, h.localizer, http.StatusUnauthorized, "invalid_access_token", "error.invalid_access_token")
	case errors.Is(err, errInvalidPlayerID):
		writeError(w, r, h.localizer, http.StatusBadRequest, "invalid_player_id", "error.invalid_player_id")
	case errors.Is(err, service.ErrEmptyTeamName):
		writeError(w, r, h.localizer, http.StatusBadRequest, "invalid_team_name", "error.invalid_team_name")
	case errors.Is(err, service.ErrEmptyPlayerFirstName), errors.Is(err, service.ErrEmptyPlayerLastName):
		writeError(w, r, h.localizer, http.StatusBadRequest, "invalid_player", "error.invalid_player")
	case errors.Is(err, repository.ErrCountryNotFound):
		writeError(w, r, h.localizer, http.StatusBadRequest, "invalid_country", "error.invalid_country")
	case errors.Is(err, repository.ErrTeamNotFound):
		writeError(w, r, h.localizer, http.StatusNotFound, "team_not_found", "error.team_not_found")
	case errors.Is(err, repository.ErrPlayerNotFound):
		writeError(w, r, h.localizer, http.StatusNotFound, "player_not_found", "error.player_not_found")
	default:
		writeError(w, r, h.localizer, http.StatusInternalServerError, "internal_error", "error.internal_error")
	}
}

var errInvalidPlayerID = errors.New("invalid player id")

// playerIDFromRequest parses and validates a player ID from the route params.
func playerIDFromRequest(r *http.Request) (int64, error) {
	playerIDParam := strings.TrimSpace(chi.URLParam(r, "playerId"))
	if playerIDParam == "" {
		return 0, errInvalidPlayerID
	}

	playerID, err := strconv.ParseInt(playerIDParam, 10, 64)
	if err != nil || playerID < 1 {
		return 0, errInvalidPlayerID
	}

	return playerID, nil
}
