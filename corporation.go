package monocle

import (
	"time"

	"github.com/volatiletech/null"
)

type Corporation struct {
	ID            uint        `db:"id" boil:"id" json:"id" toml:"id" yaml:"id"`
	Name          string      `db:"name" boil:"name" json:"name" toml:"name" yaml:"name"`
	Ticker        string      `db:"ticker" boil:"ticker" json:"ticker" toml:"ticker" yaml:"ticker"`
	MemberCount   uint        `db:"member_count" boil:"member_count" json:"member_count" toml:"member_count" yaml:"member_count"`
	CeoID         uint64      `db:"ceo_id" boil:"ceo_id" json:"ceo_id" toml:"ceo_id" yaml:"ceo_id"`
	AllianceID    null.Int    `db:"alliance_id" boil:"alliance_id" json:"alliance_id,omitempty" toml:"alliance_id" yaml:"alliance_id,omitempty"`
	DateFounded   null.Time   `db:"date_founded" boil:"date_founded" json:"date_founded,omitempty" toml:"date_founded" yaml:"date_founded,omitempty"`
	CreatorID     uint64      `db:"creator_id" boil:"creator_id" json:"creator_id" toml:"creator_id" yaml:"creator_id"`
	HomeStationID null.Uint64 `db:"home_station_id" boil:"home_station_id" json:"home_station_id,omitempty" toml:"home_station_id" yaml:"home_station_id,omitempty"`
	TaxRate       float32     `db:"tax_rate" boil:"tax_rate" json:"tax_rate" toml:"tax_rate" yaml:"tax_rate"`
	WarEligible   bool        `db:"war_eligible" boil:"war_eligible" json:"war_eligible" toml:"war_eligible" yaml:"war_eligible"`
	Ignored       bool        `db:"ignored" boil:"ignored" json:"ignored" toml:"ignored" yaml:"ignored"`
	Closed        bool        `db:"closed" boil:"closed" json:"closed" toml:"closed" yaml:"closed"`
	Etag          string      `db:"etag" boil:"etag" json:"etag" toml:"etag" yaml:"etag"`
	Expires       time.Time   `db:"expires" boil:"expires" json:"expires" toml:"expires" yaml:"expires"`
	CreatedAt     time.Time   `db:"created_at" boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`
	UpdatedAt     time.Time   `db:"updated_at" boil:"updated_at" json:"updated_at" toml:"updated_at" yaml:"updated_at"`
}

func (c Corporation) IsExpired() bool {
	return c.Expires.Before(time.Now())
}
