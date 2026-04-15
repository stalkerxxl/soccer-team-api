package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stalkerxxl/soccer-team-api/internal/auth"
	"github.com/stalkerxxl/soccer-team-api/internal/domain"
	"github.com/stalkerxxl/soccer-team-api/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthServiceSignupCreatesUserTeamAndPlayers(t *testing.T) {
	t.Parallel()

	txUserRepo := &stubAuthUserRepository{getByEmailErr: repository.ErrUserNotFound}
	teamService := &stubAuthTeamService{createdTeam: &domain.Team{ID: 101, UserID: 77}}
	playerService := &stubAuthPlayerService{}
	txManager := &stubAuthTransactionManager{
		deps: AuthTxDeps{
			UserRepo:      txUserRepo,
			TeamService:   teamService,
			PlayerService: playerService,
		},
	}
	tokenProvider := &stubAuthTokenProvider{
		accessToken:     "access-token",
		accessExpiresAt: time.Date(2026, 4, 15, 12, 0, 30, 0, time.UTC),
	}

	svc := newTestAuthService(t, &stubAuthUserRepository{}, tokenProvider, txManager)
	svc.now = func() time.Time {
		return time.Date(2026, 4, 15, 12, 0, 0, 0, time.UTC)
	}

	result, err := svc.Signup(context.Background(), SignupInput{
		Email:    "  USER@example.com ",
		Password: "strong-password",
	})
	if err != nil {
		t.Fatalf("Signup() error = %v", err)
	}

	if !txManager.called {
		t.Fatal("Signup() did not use transaction manager")
	}

	if txUserRepo.createdUser == nil {
		t.Fatal("Signup() did not create user")
	}

	if txUserRepo.createdUser.Email != "user@example.com" {
		t.Fatalf("created user email = %q, want %q", txUserRepo.createdUser.Email, "user@example.com")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(txUserRepo.createdUser.PasswordHash), []byte("strong-password")); err != nil {
		t.Fatalf("created password hash does not match password: %v", err)
	}

	if teamService.lastUserID != 77 {
		t.Fatalf("CreateDefault() user id = %d, want %d", teamService.lastUserID, 77)
	}

	if playerService.lastTeamID != 101 {
		t.Fatalf("CreateInitialSquad() team id = %d, want %d", playerService.lastTeamID, 101)
	}

	if tokenProvider.lastUserID != 77 {
		t.Fatalf("GenerateAccessToken() user id = %d, want %d", tokenProvider.lastUserID, 77)
	}

	if result.AccessToken != "access-token" {
		t.Fatalf("result.AccessToken = %q, want %q", result.AccessToken, "access-token")
	}

	if result.ExpiresIn != 30 {
		t.Fatalf("result.ExpiresIn = %d, want %d", result.ExpiresIn, 30)
	}
}

func TestAuthServiceSignupReturnsEmailTakenWhenUserExists(t *testing.T) {
	t.Parallel()

	txManager := &stubAuthTransactionManager{
		deps: AuthTxDeps{
			UserRepo:      &stubAuthUserRepository{userByEmail: &domain.User{ID: 1, Email: "user@example.com"}},
			TeamService:   &stubAuthTeamService{},
			PlayerService: &stubAuthPlayerService{},
		},
	}

	svc := newTestAuthService(t, &stubAuthUserRepository{}, &stubAuthTokenProvider{}, txManager)

	_, err := svc.Signup(context.Background(), SignupInput{Email: "user@example.com", Password: "12345"})
	if !errors.Is(err, ErrEmailAlreadyTaken) {
		t.Fatalf("Signup() error = %v, want %v", err, ErrEmailAlreadyTaken)
	}
}

func TestAuthServiceLoginRejectsInvalidCredentials(t *testing.T) {
	t.Parallel()

	passwordHash, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword() error = %v", err)
	}

	testCases := []struct {
		name     string
		userRepo *stubAuthUserRepository
		password string
	}{
		{
			name:     "user not found",
			userRepo: &stubAuthUserRepository{getByEmailErr: repository.ErrUserNotFound},
			password: "any-password",
		},
		{
			name: "wrong password",
			userRepo: &stubAuthUserRepository{userByEmail: &domain.User{
				ID: 1, Email: "user@example.com", PasswordHash: string(passwordHash),
			}},
			password: "wrong-password",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			svc := newTestAuthService(t, tc.userRepo, &stubAuthTokenProvider{}, &stubAuthTransactionManager{})

			_, err := svc.Login(context.Background(), LoginInput{Email: "user@example.com", Password: tc.password})
			if !errors.Is(err, ErrInvalidCredentials) {
				t.Fatalf("Login() error = %v, want %v", err, ErrInvalidCredentials)
			}
		})
	}
}

func TestAuthServiceLoginReturnsAccessToken(t *testing.T) {
	t.Parallel()

	passwordHash, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword() error = %v", err)
	}

	userRepo := &stubAuthUserRepository{
		userByEmail: &domain.User{ID: 15, Email: "user@example.com", PasswordHash: string(passwordHash)},
	}
	tokenProvider := &stubAuthTokenProvider{
		accessToken:     "access-token",
		accessExpiresAt: time.Date(2026, 4, 15, 12, 1, 0, 0, time.UTC),
	}
	svc := newTestAuthService(t, userRepo, tokenProvider, &stubAuthTransactionManager{})
	svc.now = func() time.Time {
		return time.Date(2026, 4, 15, 12, 0, 0, 0, time.UTC)
	}

	result, err := svc.Login(context.Background(), LoginInput{Email: " USER@example.com ", Password: "correct-password"})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	if userRepo.lastEmail != "user@example.com" {
		t.Fatalf("GetByEmail() email = %q, want %q", userRepo.lastEmail, "user@example.com")
	}

	if tokenProvider.lastUserID != 15 {
		t.Fatalf("GenerateAccessToken() user id = %d, want %d", tokenProvider.lastUserID, 15)
	}

	if result.AccessToken != "access-token" {
		t.Fatalf("result.AccessToken = %q, want %q", result.AccessToken, "access-token")
	}

	if result.ExpiresIn != 60 {
		t.Fatalf("result.ExpiresIn = %d, want %d", result.ExpiresIn, 60)
	}
}

func newTestAuthService(
	t *testing.T,
	userRepo *stubAuthUserRepository,
	tokenProvider *stubAuthTokenProvider,
	txManager *stubAuthTransactionManager,
) *AuthService {
	t.Helper()

	svc, err := NewAuthService(userRepo, tokenProvider, txManager, 5, 72)
	if err != nil {
		t.Fatalf("NewAuthService() error = %v", err)
	}

	return svc
}

type stubAuthUserRepository struct {
	userByEmail   *domain.User
	createdUser   *domain.User
	lastEmail     string
	getByEmailErr error
	createErr     error
}

func (r *stubAuthUserRepository) Create(_ context.Context, user *domain.User) error {
	if r.createErr != nil {
		return r.createErr
	}

	if user.ID == 0 {
		user.ID = 77
	}

	clone := *user
	r.createdUser = &clone
	return nil
}

func (r *stubAuthUserRepository) GetByID(_ context.Context, _ int64) (*domain.User, error) {
	return nil, nil
}

func (r *stubAuthUserRepository) GetByEmail(_ context.Context, email string) (*domain.User, error) {
	r.lastEmail = email
	if r.getByEmailErr != nil {
		return nil, r.getByEmailErr
	}
	if r.userByEmail == nil {
		return nil, repository.ErrUserNotFound
	}
	return r.userByEmail, nil
}

func (r *stubAuthUserRepository) Update(_ context.Context, _ *domain.User) error {
	return nil
}

type stubAuthTokenProvider struct {
	accessToken     string
	accessExpiresAt time.Time
	lastUserID      int64
	generateErr     error
	parseClaims     auth.TokenClaims
	parseErr        error
}

func (p *stubAuthTokenProvider) GenerateAccessToken(userID int64) (string, time.Time, error) {
	p.lastUserID = userID
	if p.generateErr != nil {
		return "", time.Time{}, p.generateErr
	}
	return p.accessToken, p.accessExpiresAt, nil
}

func (p *stubAuthTokenProvider) ParseAccessToken(_ string) (auth.TokenClaims, error) {
	return p.parseClaims, p.parseErr
}

type stubAuthTransactionManager struct {
	deps   AuthTxDeps
	called bool
	err    error
}

func (m *stubAuthTransactionManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context, deps AuthTxDeps) error) error {
	m.called = true
	if m.err != nil {
		return m.err
	}
	return fn(ctx, m.deps)
}

type stubAuthTeamService struct {
	createdTeam *domain.Team
	lastUserID  int64
	createErr   error
}

func (s *stubAuthTeamService) CreateDefault(_ context.Context, userID int64) (*domain.Team, error) {
	s.lastUserID = userID
	if s.createErr != nil {
		return nil, s.createErr
	}
	return s.createdTeam, nil
}

func (s *stubAuthTeamService) GetByID(_ context.Context, _ int64) (*domain.Team, error) {
	return nil, nil
}
func (s *stubAuthTeamService) GetAll(_ context.Context) ([]domain.Team, error) { return nil, nil }
func (s *stubAuthTeamService) GetByUserID(_ context.Context, _ int64) (*domain.Team, error) {
	return nil, nil
}
func (s *stubAuthTeamService) Update(_ context.Context, _ *domain.Team) error { return nil }
func (s *stubAuthTeamService) GetCountryByID(_ context.Context, _ int64) (*domain.Country, error) {
	return nil, nil
}
func (s *stubAuthTeamService) ListCountries(_ context.Context) ([]domain.Country, error) {
	return nil, nil
}

type stubAuthPlayerService struct {
	lastTeamID int64
	createErr  error
}

func (s *stubAuthPlayerService) CreateInitialSquad(_ context.Context, teamID int64) ([]domain.Player, error) {
	s.lastTeamID = teamID
	if s.createErr != nil {
		return nil, s.createErr
	}
	return nil, nil
}

func (s *stubAuthPlayerService) GetByID(_ context.Context, _ int64) (*domain.Player, error) {
	return nil, nil
}
func (s *stubAuthPlayerService) ListByTeamID(_ context.Context, _ int64) ([]domain.Player, error) {
	return nil, nil
}
func (s *stubAuthPlayerService) Update(_ context.Context, _ *domain.Player) error { return nil }
