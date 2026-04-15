package api

import (
	"net/http"

	"github.com/stalkerxxl/soccer-team-api/internal/auth"
	"github.com/stalkerxxl/soccer-team-api/internal/domain"
	"github.com/stalkerxxl/soccer-team-api/internal/service"
)

// currentTeam resolves the current user's team from the authenticated request context.
func currentTeam(r *http.Request, teamService service.TeamManager) (*domain.Team, error) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		return nil, service.ErrInvalidAccessToken
	}

	return teamService.GetByUserID(r.Context(), userID)
}
