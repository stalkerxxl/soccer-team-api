package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/stalkerxxl/soccer-team-api/internal/domain"
	"github.com/uptrace/bun"
)

var ErrUserNotFound = errors.New("user not found")

// UserRepository defines persistence operations for users.
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id int64) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
}

// UserRepo is a Bun-backed implementation of UserRepository.
type UserRepo struct {
	db bun.IDB
}

// NewUserRepository creates a user repository backed by the provided Bun handle.
func NewUserRepository(db bun.IDB) *UserRepo {
	return &UserRepo{db: db}
}

// Create inserts a new user.
func (r *UserRepo) Create(ctx context.Context, user *domain.User) error {
	if _, err := r.db.NewInsert().Model(user).Exec(ctx); err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	return nil
}

// GetByID returns a user by id.
func (r *UserRepo) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	user := new(domain.User)

	err := r.db.NewSelect().
		Model(user).
		Where("u.id = ?", id).
		Limit(1).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}

		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return user, nil
}

// GetByEmail returns a user by email using a case-insensitive lookup.
func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	user := new(domain.User)

	err := r.db.NewSelect().
		Model(user).
		Where("LOWER(u.email) = LOWER(?)", email).
		Limit(1).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}

		return nil, fmt.Errorf("get user by email: %w", err)
	}

	return user, nil
}

// Update persists changes to an existing user.
func (r *UserRepo) Update(ctx context.Context, user *domain.User) error {
	if _, err := r.db.NewUpdate().
		Model(user).
		WherePK().
		ExcludeColumn("created_at").
		Exec(ctx); err != nil {
		return fmt.Errorf("update user: %w", err)
	}

	return nil
}
