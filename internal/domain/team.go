package domain

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

var (
	_ bun.BeforeAppendModelHook = (*Country)(nil)
	_ bun.BeforeAppendModelHook = (*Team)(nil)
)

const (
	// DefaultTeamName is assigned to a newly created team before the user renames it.
	DefaultTeamName = "Your team"
	// DefaultCountryID is used for initial entities when no explicit country is selected yet.
	DefaultCountryID int64 = 1
	// InitialTeamBudget is the starting transfer budget for a newly created team.
	InitialTeamBudget int64 = 5_000_000
)

// Country represents a supported country in the system.
type Country struct {
	bun.BaseModel `bun:"table:countries,alias:c"`

	ID        int64     `bun:"id,pk,autoincrement" json:"id"`
	Name      string    `bun:"name,notnull" json:"name"`
	ISOCode   string    `bun:"iso_code,notnull" json:"iso_code"`
	CreatedAt time.Time `bun:"created_at,notnull,nullzero,default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time `bun:"updated_at,notnull,nullzero,default:current_timestamp" json:"updated_at"`
}

// Team represents a user's soccer team.
type Team struct {
	bun.BaseModel `bun:"table:teams,alias:t"`

	ID                      int64     `bun:"id,pk,autoincrement" json:"id"`
	UserID                  int64     `bun:"user_id,notnull" json:"user_id"`
	CountryID               int64     `bun:"country_id,notnull" json:"country_id"`
	Name                    string    `bun:"name,notnull" json:"name"`
	Budget                  int64     `bun:"budget,notnull" json:"budget"`
	TotalPlayersMarketValue int64     `bun:"total_players_market_value,scanonly" json:"total_players_market_value"`
	CreatedAt               time.Time `bun:"created_at,notnull,nullzero,default:current_timestamp" json:"created_at"`
	UpdatedAt               time.Time `bun:"updated_at,notnull,nullzero,default:current_timestamp" json:"updated_at"`

	User    *User    `bun:"rel:belongs-to,join:user_id=id" json:"user,omitempty"`
	Country *Country `bun:"rel:belongs-to,join:country_id=id" json:"country,omitempty"`
}

// BeforeAppendModel keeps Bun-managed timestamps in sync for inserts and updates.
func (c *Country) BeforeAppendModel(_ context.Context, query bun.Query) error {
	applyCreatedUpdatedTimestamps(query, &c.CreatedAt, &c.UpdatedAt)
	return nil
}

// BeforeAppendModel keeps Bun-managed timestamps in sync for inserts and updates.
func (t *Team) BeforeAppendModel(_ context.Context, query bun.Query) error {
	applyCreatedUpdatedTimestamps(query, &t.CreatedAt, &t.UpdatedAt)
	return nil
}
