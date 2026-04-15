package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stalkerxxl/soccer-team-api/internal/domain"
)

func TestUserServiceNormalizeEmail(t *testing.T) {
	t.Parallel()

	got := NormalizeEmail("  USER@Example.COM ")

	if got != "user@example.com" {
		t.Fatalf("NormalizeEmail() = %q, want %q", got, "user@example.com")
	}
}

func TestUserServiceValidateEmail(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		email   string
		wantErr error
	}{
		{name: "valid", email: "user@example.com"},
		{name: "empty", email: "   ", wantErr: ErrEmptyEmail},
		{name: "invalid", email: "not-an-email", wantErr: ErrInvalidEmail},
		{name: "display name", email: "Roman <user@example.com>", wantErr: ErrInvalidEmail},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateEmail(tc.email)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("ValidateEmail() error = %v, want %v", err, tc.wantErr)
			}
		})
	}
}

func TestUserServiceValidatePassword(t *testing.T) {
	t.Parallel()

	const minLen = 5
	const maxLen = 72

	testCases := []struct {
		name     string
		password string
		wantErr  error
	}{
		{name: "valid ascii", password: "12345"},
		{name: "valid unicode", password: "აბგდე"},
		{name: "too short ascii", password: "1234", wantErr: ErrPasswordTooShort},
		{name: "too short unicode by characters", password: "აბგდ", wantErr: ErrPasswordTooShort},
		{name: "too long", password: strings.Repeat("a", 73), wantErr: ErrPasswordTooLong},
		{name: "too long unicode by bytes", password: strings.Repeat("ა", 25), wantErr: ErrPasswordTooLong},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := ValidatePassword(tc.password, minLen, maxLen)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("ValidatePassword() error = %v, want %v", err, tc.wantErr)
			}
		})
	}
}

func TestUserServiceGetByEmailNormalizesInput(t *testing.T) {
	t.Parallel()

	repo := &stubUserRepository{
		userByEmail: &domain.User{ID: 1, Email: "user@example.com"},
	}
	svc := newTestUserService(t, repo)

	user, err := svc.GetByEmail(context.Background(), "  USER@example.com ")
	if err != nil {
		t.Fatalf("GetByEmail() error = %v", err)
	}

	if user == nil {
		t.Fatal("GetByEmail() returned nil user")
	}

	if repo.lastEmail != "user@example.com" {
		t.Fatalf("repository email = %q, want %q", repo.lastEmail, "user@example.com")
	}
}

func newTestUserService(t *testing.T, userRepo *stubUserRepository) UserManager {
	t.Helper()

	svc, err := NewUserService(userRepo)
	if err != nil {
		t.Fatalf("NewUserService() error = %v", err)
	}

	return svc
}

type stubUserRepository struct {
	userByID    *domain.User
	userByEmail *domain.User
	lastEmail   string
}

func (r *stubUserRepository) Create(_ context.Context, _ *domain.User) error {
	return nil
}

func (r *stubUserRepository) GetByID(_ context.Context, _ int64) (*domain.User, error) {
	return r.userByID, nil
}

func (r *stubUserRepository) GetByEmail(_ context.Context, email string) (*domain.User, error) {
	r.lastEmail = email
	return r.userByEmail, nil
}

func (r *stubUserRepository) Update(_ context.Context, _ *domain.User) error {
	return nil
}
