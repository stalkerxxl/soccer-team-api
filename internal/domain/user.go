package domain

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

var _ bun.BeforeAppendModelHook = (*User)(nil)

// User represents an authenticated API user.
type User struct {
	bun.BaseModel `bun:"table:users,alias:u"`

	ID           int64     `bun:"id,pk,autoincrement" json:"id"`
	Email        string    `bun:"email,notnull" json:"email"`
	PasswordHash string    `bun:"password_hash,notnull" json:"-"`
	CreatedAt    time.Time `bun:"created_at,notnull,nullzero,default:current_timestamp" json:"created_at"`
	UpdatedAt    time.Time `bun:"updated_at,notnull,nullzero,default:current_timestamp" json:"updated_at"`
}

// BeforeAppendModel keeps Bun-managed timestamps in sync for inserts and updates.
func (u *User) BeforeAppendModel(_ context.Context, query bun.Query) error {
	applyCreatedUpdatedTimestamps(query, &u.CreatedAt, &u.UpdatedAt)
	return nil
}
