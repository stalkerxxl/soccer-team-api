package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stalkerxxl/soccer-team-api/internal/service"
)

func TestAuthHandlerLocalizesInvalidEmailToGeorgian(t *testing.T) {
	t.Parallel()

	authService := &stubAPIAuthService{signupErr: service.ErrInvalidEmail}
	router := NewRouter(NewAuthHandler(authService, apiTestLocalizer), nil, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/auth/signup", bytes.NewBufferString(`{"email":"bad","password":"secret123"}`))
	req.Header.Set("Accept-Language", "ka")
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

	if response.Code != "invalid_email" {
		t.Fatalf("response.Code = %q, want %q", response.Code, "invalid_email")
	}

	if response.Message != "ელფოსტის მისამართი არასწორია." {
		t.Fatalf("response.Message = %q, want %q", response.Message, "ელფოსტის მისამართი არასწორია.")
	}
}

func TestAuthHandlerFallsBackToEnglishForUnsupportedLanguage(t *testing.T) {
	t.Parallel()

	authService := &stubAPIAuthService{loginErr: service.ErrInvalidCredentials}
	router := NewRouter(NewAuthHandler(authService, apiTestLocalizer), nil, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(`{"email":"user@example.com","password":"wrong"}`))
	req.Header.Set("Accept-Language", "fr")
	req.Header.Set("Content-Type", "application/json")
	rrw := httptest.NewRecorder()

	router.ServeHTTP(rrw, req)

	if rrw.Code != http.StatusUnauthorized {
		t.Fatalf("status code = %d, want %d", rrw.Code, http.StatusUnauthorized)
	}

	var response errorResponse
	if err := json.Unmarshal(rrw.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if response.Code != "invalid_credentials" {
		t.Fatalf("response.Code = %q, want %q", response.Code, "invalid_credentials")
	}

	if response.Message != "Invalid credentials." {
		t.Fatalf("response.Message = %q, want %q", response.Message, "Invalid credentials.")
	}
}

type stubAPIAuthService struct {
	signupResponse *service.AuthTokens
	loginResponse  *service.AuthTokens
	signupErr      error
	loginErr       error
}

func (s *stubAPIAuthService) Signup(_ context.Context, _ service.SignupInput) (*service.AuthTokens, error) {
	if s.signupErr != nil {
		return nil, s.signupErr
	}

	if s.signupResponse != nil {
		return s.signupResponse, nil
	}

	return &service.AuthTokens{}, nil
}

func (s *stubAPIAuthService) Login(_ context.Context, _ service.LoginInput) (*service.AuthTokens, error) {
	if s.loginErr != nil {
		return nil, s.loginErr
	}

	if s.loginResponse != nil {
		return s.loginResponse, nil
	}

	return &service.AuthTokens{}, nil
}
