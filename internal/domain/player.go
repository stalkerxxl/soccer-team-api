package domain

import (
	"context"
	"slices"
	"time"

	"github.com/uptrace/bun"
)

const (
	// InitialPlayerMarketValue is assigned to newly created players.
	InitialPlayerMarketValue int64 = 1_000_000
	// InitialGoalkeepers is the number of goalkeepers in the starter squad.
	InitialGoalkeepers = 3
	// InitialDefenders is the number of defenders in the starter squad.
	InitialDefenders = 6
	// InitialMidfielders is the number of midfielders in the starter squad.
	InitialMidfielders = 6
	// InitialAttackers is the number of attackers in the starter squad.
	InitialAttackers = 5
	// InitialSquadSize is the total number of players in the starter squad.
	InitialSquadSize = InitialGoalkeepers + InitialDefenders + InitialMidfielders + InitialAttackers
	// MinPlayerAge is the lower age bound accepted by the domain.
	MinPlayerAge = 18
	// MaxPlayerAge is the upper age bound accepted by the domain.
	MaxPlayerAge = 40
)

// PlayerPosition defines the role of a player on the field.
type PlayerPosition string

const (
	// PlayerPositionGoalkeeper identifies a goalkeeper.
	PlayerPositionGoalkeeper PlayerPosition = "goalkeeper"
	// PlayerPositionDefender identifies a defender.
	PlayerPositionDefender PlayerPosition = "defender"
	// PlayerPositionMidfielder identifies a midfielder.
	PlayerPositionMidfielder PlayerPosition = "midfielder"
	// PlayerPositionAttacker identifies an attacker.
	PlayerPositionAttacker PlayerPosition = "attacker"
)

var (
	_ bun.BeforeAppendModelHook = (*Player)(nil)

	// ValidPlayerPositions enumerates all supported player positions.
	ValidPlayerPositions = []PlayerPosition{
		PlayerPositionGoalkeeper,
		PlayerPositionDefender,
		PlayerPositionMidfielder,
		PlayerPositionAttacker,
	}
	// InitialSquadComposition defines how many players of each position are created initially.
	InitialSquadComposition = map[PlayerPosition]int{
		PlayerPositionGoalkeeper: InitialGoalkeepers,
		PlayerPositionDefender:   InitialDefenders,
		PlayerPositionMidfielder: InitialMidfielders,
		PlayerPositionAttacker:   InitialAttackers,
	}
)

// Player represents a football player owned by a team.
type Player struct {
	bun.BaseModel `bun:"table:players,alias:p"`

	ID          int64          `bun:"id,pk,autoincrement" json:"id"`
	TeamID      int64          `bun:"team_id,notnull" json:"team_id"`
	FirstName   string         `bun:"first_name,notnull" json:"first_name"`
	LastName    string         `bun:"last_name,notnull" json:"last_name"`
	CountryID   int64          `bun:"country_id,notnull" json:"country_id"`
	Age         int            `bun:"age,notnull" json:"age"`
	Position    PlayerPosition `bun:"position,notnull,type:text" json:"position"`
	MarketValue int64          `bun:"market_value,notnull" json:"market_value"`
	CreatedAt   time.Time      `bun:"created_at,notnull,nullzero,default:current_timestamp" json:"created_at"`
	UpdatedAt   time.Time      `bun:"updated_at,notnull,nullzero,default:current_timestamp" json:"updated_at"`

	Team    *Team    `bun:"rel:belongs-to,join:team_id=id" json:"team,omitempty"`
	Country *Country `bun:"rel:belongs-to,join:country_id=id" json:"country,omitempty"`
}

// BeforeAppendModel keeps Bun-managed timestamps in sync for inserts and updates.
func (p *Player) BeforeAppendModel(_ context.Context, query bun.Query) error {
	applyCreatedUpdatedTimestamps(query, &p.CreatedAt, &p.UpdatedAt)
	return nil
}

// IsValid reports whether p is one of the supported player positions.
func (p PlayerPosition) IsValid() bool {
	return slices.Contains(ValidPlayerPositions, p)
}
