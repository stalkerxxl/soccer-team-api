package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// NewRouter wires all HTTP routes and applies auth middleware to protected groups.
func NewRouter(
	authHandler *AuthHandler,
	meHandler *MeHandler,
	marketHandler *MarketHandler,
	authMiddleware func(http.Handler) http.Handler,
) http.Handler {
	router := chi.NewRouter()

	router.Get("/health", healthHandler)
	if authHandler != nil {
		router.Post("/auth/signup", authHandler.Signup)
		router.Post("/auth/login", authHandler.Login)
	}
	if meHandler != nil && authMiddleware != nil {
		router.Route("/me", func(r chi.Router) {
			r.Use(authMiddleware)
			r.Get("/team", meHandler.GetTeam)
			r.Patch("/team", meHandler.UpdateTeam)
			r.Get("/players", meHandler.ListPlayers)
			r.Get("/players/{playerId}", meHandler.GetPlayer)
			r.Patch("/players/{playerId}", meHandler.UpdatePlayer)
		})
	}
	if marketHandler != nil {
		router.Route("/market", func(r chi.Router) {
			r.Get("/players", marketHandler.ListPlayers)
			if authMiddleware != nil {
				r.Group(func(r chi.Router) {
					r.Use(authMiddleware)
					r.Post("/players/{playerId}/list", marketHandler.CreateListing)
					r.Delete("/players/{playerId}/list", marketHandler.CancelListing)
					r.Post("/listings/{listingId}/buy", marketHandler.BuyListing)
				})
			}
		})
	}

	return router
}

// healthHandler reports a minimal liveness response for basic health checks.
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
