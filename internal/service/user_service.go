package service

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"unicode/utf8"

	"github.com/stalkerxxl/soccer-team-api/internal/domain"
	"github.com/stalkerxxl/soccer-team-api/internal/repository"
)

var (
	// ErrInvalidEmail indicates that the email does not match the expected format.
	ErrInvalidEmail = errors.New("invalid email format")
	// ErrPasswordTooShort indicates that the password is shorter than the configured minimum.
	ErrPasswordTooShort = errors.New("password is too short")
	// ErrPasswordTooLong indicates that the password exceeds the configured maximum byte length.
	ErrPasswordTooLong = errors.New("password is too long")
	// ErrEmptyEmail indicates that the email is required.
	ErrEmptyEmail = errors.New("email is required")
	// ErrUserRepositoryNil indicates that the service requires a user repository.
	ErrUserRepositoryNil = errors.New("user repository is required")
)

// UserManager defines read-only user use cases exposed to other services.
type UserManager interface {
	GetByID(ctx context.Context, id int64) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
}

// UserService implements user lookup and validation helpers.
type UserService struct {
	userRepo repository.UserRepository
}

// NewUserService creates a user service backed by the provided repository.
func NewUserService(userRepo repository.UserRepository) (*UserService, error) {
	if userRepo == nil {
		return nil, ErrUserRepositoryNil
	}

	return &UserService{
		userRepo: userRepo,
	}, nil
}

// GetByID returns a user by id.
func (s *UserService) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

// GetByEmail normalizes and validates the email before loading the user.
func (s *UserService) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	normalizedEmail := NormalizeEmail(email)
	if err := ValidateEmail(normalizedEmail); err != nil {
		return nil, err
	}

	return s.userRepo.GetByEmail(ctx, normalizedEmail)
}

// NormalizeEmail trims surrounding whitespace and lowercases the email.
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// ValidateEmail checks that the normalized email is non-empty and syntactically valid.
func ValidateEmail(email string) error {
	normalizedEmail := NormalizeEmail(email)
	if normalizedEmail == "" {
		return ErrEmptyEmail
	}

	parsedAddress, err := mail.ParseAddress(normalizedEmail)
	if err != nil {
		return ErrInvalidEmail
	}

	if parsedAddress.Address != normalizedEmail {
		return ErrInvalidEmail
	}

	return nil
}

// ValidatePassword enforces the configured password length limits.
// The minimum is counted in characters, while the maximum is counted in bytes
// to stay aligned with bcrypt input constraints.
func ValidatePassword(password string, minLen, maxLen int) error {
	passwordCharsLen := utf8.RuneCountInString(password)
	passwordBytesLen := len(password)

	if passwordCharsLen < minLen {
		return fmt.Errorf("%w: minimum is %d characters", ErrPasswordTooShort, minLen)
	}

	if passwordBytesLen > maxLen {
		return fmt.Errorf("%w: maximum is %d bytes", ErrPasswordTooLong, maxLen)
	}

	return nil
}
