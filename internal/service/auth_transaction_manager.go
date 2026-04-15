package service

import (
	"context"
	"errors"

	dbtx "github.com/stalkerxxl/soccer-team-api/internal/db"
	"github.com/stalkerxxl/soccer-team-api/internal/repository"
	"github.com/uptrace/bun"
)

// ErrAuthTransactionManagerNil indicates that the auth service requires a transaction manager.
var ErrAuthTransactionManagerNil = errors.New("auth transaction manager is required")

// AuthTxDeps groups transaction-scoped dependencies used during signup.
type AuthTxDeps struct {
	UserRepo      repository.UserRepository
	TeamService   TeamManager
	PlayerService PlayerManager
}

// AuthTransactionManager executes auth workflows inside a transaction and injects tx-scoped dependencies.
type AuthTransactionManager interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context, deps AuthTxDeps) error) error
}

// authTransactionManager is the default AuthTransactionManager implementation.
type authTransactionManager struct {
	txManager dbtx.TxManager
}

// NewAuthTransactionManager creates an auth transaction manager backed by the provided DB transaction manager.
func NewAuthTransactionManager(txManager dbtx.TxManager) (AuthTransactionManager, error) {
	if txManager == nil {
		return nil, ErrAuthTransactionManagerNil
	}

	return &authTransactionManager{txManager: txManager}, nil
}

// WithinTransaction runs fn inside a database transaction with repositories and services bound to txDB.
func (m *authTransactionManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context, deps AuthTxDeps) error) error {
	return m.txManager.WithinTransaction(ctx, func(ctx context.Context, txDB bun.IDB) error {
		// Each dependency must be recreated on top of txDB so the whole signup flow shares the same transaction.
		userRepo := repository.NewUserRepository(txDB)
		teamRepo := repository.NewTeamRepository(txDB)
		playerRepo := repository.NewPlayerRepository(txDB)

		teamService, err := NewTeamService(teamRepo)
		if err != nil {
			return err
		}

		playerService, err := NewPlayerService(playerRepo, teamRepo)
		if err != nil {
			return err
		}

		return fn(ctx, AuthTxDeps{
			UserRepo:      userRepo,
			TeamService:   teamService,
			PlayerService: playerService,
		})
	})
}
