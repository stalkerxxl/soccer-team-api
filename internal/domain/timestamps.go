package domain

import (
	"time"

	"github.com/uptrace/bun"
)

// applyCreatedUpdatedTimestamps updates audit timestamps for models that track both
// creation and modification times.
func applyCreatedUpdatedTimestamps(query bun.Query, createdAt, updatedAt *time.Time) {
	now := time.Now().UTC()

	switch query.(type) {
	case *bun.InsertQuery:
		// Preserve an explicitly prefilled created_at while always refreshing updated_at.
		if createdAt.IsZero() {
			*createdAt = now
		}
		*updatedAt = now
	case *bun.UpdateQuery:
		*updatedAt = now
	}
}

// applyCreatedAtOnInsert initializes created_at for models that do not store updated_at.
func applyCreatedAtOnInsert(query bun.Query, createdAt *time.Time) {
	if _, ok := query.(*bun.InsertQuery); ok && createdAt.IsZero() {
		*createdAt = time.Now().UTC()
	}
}
