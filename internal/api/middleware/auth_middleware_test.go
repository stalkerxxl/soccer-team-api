package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stalkerxxl/soccer-team-api/internal/auth"
	"github.com/stalkerxxl/soccer-team-api/internal/i18n"
)

var middlewareTestLocalizer = i18n.MustNewLocalizer()

func TestAuthMiddlewareAddsUserIDToContext(t *testing.T) {
	t.Parallel()

	provider := &stubMiddlewareTokenProvider{
		claims: auth.TokenClaims{UserID: 42},
	}

	handler := Auth(provider, middlewareTestLocalizer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := auth.UserIDFromContext(r.Context())
		if !ok {
			t.Fatal("user id not found in context")
		}
		if userID != 42 {
			t.Fatalf("userID = %d, want %d", userID, 42)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/me/team", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rrw := httptest.NewRecorder()

	handler.ServeHTTP(rrw, req)

	if rrw.Code != http.StatusNoContent {
		t.Fatalf("status code = %d, want %d", rrw.Code, http.StatusNoContent)
	}
}

func TestAuthMiddlewareRejectsMissingToken(t *testing.T) {
	t.Parallel()

	handler := Auth(&stubMiddlewareTokenProvider{}, middlewareTestLocalizer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/me/team", nil)
	rrw := httptest.NewRecorder()

	handler.ServeHTTP(rrw, req)

	assertUnauthorizedResponse(t, rrw, "invalid_access_token", "Invalid access token.")
}

func TestAuthMiddlewareRejectsInvalidToken(t *testing.T) {
	t.Parallel()

	handler := Auth(&stubMiddlewareTokenProvider{parseErr: auth.ErrInvalidToken}, middlewareTestLocalizer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/me/team", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rrw := httptest.NewRecorder()

	handler.ServeHTTP(rrw, req)

	assertUnauthorizedResponse(t, rrw, "invalid_access_token", "Invalid access token.")
}

func TestAuthMiddlewareLocalizesUnauthorizedToGeorgian(t *testing.T) {
	t.Parallel()

	handler := Auth(&stubMiddlewareTokenProvider{}, middlewareTestLocalizer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/me/team", nil)
	req.Header.Set("Accept-Language", "ka")
	rrw := httptest.NewRecorder()

	handler.ServeHTTP(rrw, req)

	assertUnauthorizedResponse(t, rrw, "invalid_access_token", "წვდომის ტოკენი არასწორია.")
}

func assertUnauthorizedResponse(t *testing.T, rrw *httptest.ResponseRecorder, wantCode string, wantMessage string) {
	t.Helper()

	if rrw.Code != http.StatusUnauthorized {
		t.Fatalf("status code = %d, want %d", rrw.Code, http.StatusUnauthorized)
	}

	var response errorResponse
	if err := json.Unmarshal(rrw.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if response.Code != wantCode {
		t.Fatalf("response.Code = %q, want %q", response.Code, wantCode)
	}

	if response.Message != wantMessage {
		t.Fatalf("response.Message = %q, want %q", response.Message, wantMessage)
	}
}

type stubMiddlewareTokenProvider struct {
	claims   auth.TokenClaims
	parseErr error
}

func (p *stubMiddlewareTokenProvider) GenerateAccessToken(_ int64) (string, time.Time, error) {
	return "", time.Time{}, nil
}

func (p *stubMiddlewareTokenProvider) ParseAccessToken(_ string) (auth.TokenClaims, error) {
	return p.claims, p.parseErr
}
