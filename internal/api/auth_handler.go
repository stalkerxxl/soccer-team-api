package api

import (
	"errors"
	"net/http"

	"github.com/stalkerxxl/soccer-team-api/internal/i18n"
	"github.com/stalkerxxl/soccer-team-api/internal/service"
)

type AuthHandler struct {
	authService service.AuthManager
	localizer   *i18n.Localizer
}

type signupRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// NewAuthHandler creates an auth HTTP handler with localized error responses.
func NewAuthHandler(authService service.AuthManager, localizer *i18n.Localizer) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		localizer:   localizer,
	}
}

// Signup creates a new user account and returns access tokens.
func (h *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	var req signupRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, h.localizer, http.StatusBadRequest, "invalid_request", "error.invalid_request")
		return
	}

	response, err := h.authService.Signup(r.Context(), service.SignupInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		h.writeAuthError(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, response)
}

// Login authenticates a user and returns access tokens.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, h.localizer, http.StatusBadRequest, "invalid_request", "error.invalid_request")
		return
	}

	response, err := h.authService.Login(r.Context(), service.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		h.writeAuthError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, response)
}

// writeAuthError maps auth and validation errors to API responses.
func (h *AuthHandler) writeAuthError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, service.ErrEmailAlreadyTaken):
		writeError(w, r, h.localizer, http.StatusConflict, "email_taken", "error.email_taken")
	case errors.Is(err, service.ErrInvalidCredentials):
		writeError(w, r, h.localizer, http.StatusUnauthorized, "invalid_credentials", "error.invalid_credentials")
	case errors.Is(err, service.ErrInvalidAccessToken):
		writeError(w, r, h.localizer, http.StatusUnauthorized, "invalid_access_token", "error.invalid_access_token")
	case errors.Is(err, service.ErrInvalidEmail), errors.Is(err, service.ErrEmptyEmail):
		writeError(w, r, h.localizer, http.StatusBadRequest, "invalid_email", "error.invalid_email")
	case errors.Is(err, service.ErrPasswordTooShort), errors.Is(err, service.ErrPasswordTooLong):
		writeError(w, r, h.localizer, http.StatusBadRequest, "invalid_password", "error.invalid_password")
	default:
		writeError(w, r, h.localizer, http.StatusInternalServerError, "internal_error", "error.internal_error")
	}
}
