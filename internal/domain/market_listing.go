package domain

import (
	"context"
	"slices"
	"time"

	"github.com/uptrace/bun"
)

var _ bun.BeforeAppendModelHook = (*MarketListing)(nil)

// MarketListingStatus describes the lifecycle state of a market listing.
type MarketListingStatus string

const (
	// MarketListingStatusActive marks a listing that is currently available for purchase.
	MarketListingStatusActive MarketListingStatus = "active"
	// MarketListingStatusSold marks a listing that has been purchased.
	MarketListingStatusSold MarketListingStatus = "sold"
	// MarketListingStatusCancelled marks a listing that was removed from the market.
	MarketListingStatusCancelled MarketListingStatus = "cancelled"
)

// ValidMarketListingStatuses enumerates all supported market listing statuses.
var ValidMarketListingStatuses = []MarketListingStatus{
	MarketListingStatusActive,
	MarketListingStatusSold,
	MarketListingStatusCancelled,
}

// MarketListing represents a player listed on the transfer market.
type MarketListing struct {
	bun.BaseModel `bun:"table:market_listings,alias:ml"`

	ID           int64               `bun:"id,pk,autoincrement" json:"id"`
	PlayerID     int64               `bun:"player_id,notnull" json:"player_id"`
	SellerTeamID int64               `bun:"seller_team_id,notnull" json:"seller_team_id"`
	AskingPrice  int64               `bun:"asking_price,notnull" json:"asking_price"`
	Status       MarketListingStatus `bun:"status,notnull,type:text" json:"status"`
	CreatedAt    time.Time           `bun:"created_at,notnull,nullzero,default:current_timestamp" json:"created_at"`
	UpdatedAt    time.Time           `bun:"updated_at,notnull,nullzero,default:current_timestamp" json:"updated_at"`

	Player     *Player `bun:"rel:belongs-to,join:player_id=id" json:"player,omitempty"`
	SellerTeam *Team   `bun:"rel:belongs-to,join:seller_team_id=id" json:"seller_team,omitempty"`
}

// BeforeAppendModel keeps Bun-managed timestamps in sync for inserts and updates.
func (l *MarketListing) BeforeAppendModel(_ context.Context, query bun.Query) error {
	applyCreatedUpdatedTimestamps(query, &l.CreatedAt, &l.UpdatedAt)
	return nil
}

// IsValid reports whether s is one of the supported listing statuses.
func (s MarketListingStatus) IsValid() bool {
	return slices.Contains(ValidMarketListingStatuses, s)
}
