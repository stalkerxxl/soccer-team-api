package middleware

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/stalkerxxl/soccer-team-api/internal/auth"
	"github.com/stalkerxxl/soccer-team-api/internal/i18n"
)

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Auth validates a bearer access token and stores the user ID in the request context.
func Auth(tokenProvider auth.TokenProvider, localizer *i18n.Localizer) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString, ok := bearerTokenFromHeader(r.Header.Get("Authorization"))
			if !ok {
				writeUnauthorized(w, r, localizer, "invalid_access_token", "error.invalid_access_token")
				return
			}

			claims, err := tokenProvider.ParseAccessToken(tokenString)
			if err != nil {
				writeUnauthorized(w, r, localizer, "invalid_access_token", "error.invalid_access_token")
				return
			}

			// JWT validation is sufficient for now; a DB existence check can be added later if needed.
			ctx := auth.ContextWithUserID(r.Context(), claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// bearerTokenFromHeader extracts a bearer token from the Authorization header value.
func bearerTokenFromHeader(header string) (string, bool) {
	header = strings.TrimSpace(header)
	if header == "" {
		return "", false
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", false
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", false
	}

	return token, true
}

// writeUnauthorized writes a localized unauthorized error response.
func writeUnauthorized(w http.ResponseWriter, r *http.Request, localizer *i18n.Localizer, code string, messageKey string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(errorResponse{
		Code:    code,
		Message: localizer.Message(r.Header.Get("Accept-Language"), messageKey),
	})
}
