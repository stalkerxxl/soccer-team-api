package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/stalkerxxl/soccer-team-api/internal/auth"
	"github.com/stalkerxxl/soccer-team-api/internal/domain"
	"github.com/stalkerxxl/soccer-team-api/internal/repository"
	"github.com/uptrace/bun/driver/pgdriver"
	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrTokenProviderNil indicates that the auth service requires a token provider dependency.
	ErrTokenProviderNil = errors.New("token provider is required")
	// ErrEmailAlreadyTaken indicates that a user with the same email already exists.
	ErrEmailAlreadyTaken = errors.New("email is already taken")
	// ErrInvalidCredentials indicates that the provided login credentials are invalid.
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrInvalidAccessToken indicates that the provided access token is invalid.
	ErrInvalidAccessToken = errors.New("invalid access token")
)

// AuthManager defines authentication use cases exposed to the HTTP layer.
type AuthManager interface {
	Signup(ctx context.Context, input SignupInput) (*AuthTokens, error)
	Login(ctx context.Context, input LoginInput) (*AuthTokens, error)
}

// SignupInput contains user credentials required to register a new account.
type SignupInput struct {
	Email    string
	Password string
}

// LoginInput contains user credentials required to sign in.
type LoginInput struct {
	Email    string
	Password string
}

// AuthTokens contains the access token issued after successful authentication.
type AuthTokens struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

// AuthService coordinates signup and login flows.
type AuthService struct {
	userRepo       repository.UserRepository
	tokenProvider  auth.TokenProvider
	txManager      AuthTransactionManager
	passwordMinLen int
	passwordMaxLen int
	now            func() time.Time
}

// NewAuthService creates an auth service with its required dependencies and password constraints.
func NewAuthService(
	userRepo repository.UserRepository,
	tokenProvider auth.TokenProvider,
	txManager AuthTransactionManager,
	passwordMinLen int,
	passwordMaxLen int,
) (*AuthService, error) {
	if userRepo == nil {
		return nil, ErrUserRepositoryNil
	}

	if tokenProvider == nil {
		return nil, ErrTokenProviderNil
	}

	if txManager == nil {
		return nil, ErrAuthTransactionManagerNil
	}

	if passwordMinLen < 1 {
		return nil, errors.New("password min len must be greater than 0")
	}

	if passwordMaxLen < passwordMinLen {
		return nil, errors.New("password max len must be greater than or equal to min len")
	}

	return &AuthService{
		userRepo:       userRepo,
		tokenProvider:  tokenProvider,
		txManager:      txManager,
		passwordMinLen: passwordMinLen,
		passwordMaxLen: passwordMaxLen,
		now:            time.Now,
	}, nil
}

// Signup creates a user, provisions the default team and squad, and returns an access token.
func (s *AuthService) Signup(ctx context.Context, input SignupInput) (*AuthTokens, error) {
	normalizedEmail := NormalizeEmail(input.Email)
	if err := ValidateEmail(normalizedEmail); err != nil {
		return nil, err
	}

	if err := ValidatePassword(input.Password, s.passwordMinLen, s.passwordMaxLen); err != nil {
		return nil, err
	}

	var userID int64

	err := s.txManager.WithinTransaction(ctx, func(ctx context.Context, deps AuthTxDeps) error {
		if _, err := deps.UserRepo.GetByEmail(ctx, normalizedEmail); err == nil {
			return ErrEmailAlreadyTaken
		} else if !errors.Is(err, repository.ErrUserNotFound) {
			return err
		}

		passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("hash password: %w", err)
		}

		user := &domain.User{
			Email:        normalizedEmail,
			PasswordHash: string(passwordHash),
		}
		if err := deps.UserRepo.Create(ctx, user); err != nil {
			if isUsersEmailUniqueViolation(err) {
				return ErrEmailAlreadyTaken
			}
			return err
		}

		team, err := deps.TeamService.CreateDefault(ctx, user.ID)
		if err != nil {
			return err
		}

		if _, err := deps.PlayerService.CreateInitialSquad(ctx, team.ID); err != nil {
			return err
		}

		userID = user.ID
		return nil
	})
	if err != nil {
		return nil, err
	}

	return s.issueAccessToken(userID)
}

// Login validates user credentials and returns a fresh access token.
func (s *AuthService) Login(ctx context.Context, input LoginInput) (*AuthTokens, error) {
	normalizedEmail := NormalizeEmail(input.Email)
	if err := ValidateEmail(normalizedEmail); err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByEmail(ctx, normalizedEmail)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("compare password: %w", err)
	}

	return s.issueAccessToken(user.ID)
}

// issueAccessToken generates an access token and converts its expiry into a relative lifetime.
func (s *AuthService) issueAccessToken(userID int64) (*AuthTokens, error) {
	accessToken, expiresAt, err := s.tokenProvider.GenerateAccessToken(userID)
	if err != nil {
		return nil, err
	}

	return &AuthTokens{
		AccessToken: accessToken,
		ExpiresIn:   secondsUntil(s.now().UTC(), expiresAt),
	}, nil
}

// secondsUntil returns the whole-second lifetime remaining until expiresAt.
func secondsUntil(now, expiresAt time.Time) int64 {
	if !expiresAt.After(now) {
		return 0
	}

	return int64(expiresAt.Sub(now).Seconds())
}

// isUsersEmailUniqueViolation maps the database unique constraint to a stable domain error.
func isUsersEmailUniqueViolation(err error) bool {
	var pgErr pgdriver.Error
	if !errors.As(err, &pgErr) {
		return false
	}

	return pgErr.IntegrityViolation() && pgErr.Field('n') == "users_email_lower_uq"
}
