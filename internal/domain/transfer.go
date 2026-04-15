package domain

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

const (
	// MinTransferValueIncreasePercent is the lower bound for post-transfer market value growth.
	MinTransferValueIncreasePercent = 10
	// MaxTransferValueIncreasePercent is the upper bound for post-transfer market value growth.
	MaxTransferValueIncreasePercent = 100
)

var _ bun.BeforeAppendModelHook = (*Transfer)(nil)

// Transfer records a completed player purchase between two teams.
type Transfer struct {
	bun.BaseModel `bun:"table:transfers,alias:tr"`

	ID             int64     `bun:"id,pk,autoincrement" json:"id"`
	PlayerID       int64     `bun:"player_id,notnull" json:"player_id"`
	SellerTeamID   int64     `bun:"seller_team_id,notnull" json:"seller_team_id"`
	BuyerTeamID    int64     `bun:"buyer_team_id,notnull" json:"buyer_team_id"`
	Price          int64     `bun:"price,notnull" json:"price"`
	OldMarketValue int64     `bun:"old_market_value,notnull" json:"old_market_value"`
	NewMarketValue int64     `bun:"new_market_value,notnull" json:"new_market_value"`
	CreatedAt      time.Time `bun:"created_at,notnull,nullzero,default:current_timestamp" json:"created_at"`

	Player     *Player `bun:"rel:belongs-to,join:player_id=id" json:"player,omitempty"`
	SellerTeam *Team   `bun:"rel:belongs-to,join:seller_team_id=id" json:"seller_team,omitempty"`
	BuyerTeam  *Team   `bun:"rel:belongs-to,join:buyer_team_id=id" json:"buyer_team,omitempty"`
}

// BeforeAppendModel initializes the creation timestamp when the transfer is inserted.
func (t *Transfer) BeforeAppendModel(_ context.Context, query bun.Query) error {
	applyCreatedAtOnInsert(query, &t.CreatedAt)
	return nil
}
