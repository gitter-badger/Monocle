package monocle

import (
	"time"

	"github.com/volatiletech/null"
)

type Character struct {
	ID             uint64    `db:"id" boil:"id" json:"id" toml:"id" yaml:"id"`
	Name           string    `db:"name" boil:"name" json:"name" toml:"name" yaml:"name"`
	Birthday       null.Time `db:"birthday" boil:"birthday" json:"birthday,omitempty" toml:"birthday" yaml:"birthday,omitempty"`
	Gender         string    `db:"gender" boil:"gender" json:"gender" toml:"gender" yaml:"gender"`
	SecurityStatus float32   `db:"security_status" boil:"security_status" json:"security_status" toml:"security_status" yaml:"security_status"`
	AllianceID     null.Uint `db:"alliance_id" boil:"alliance_id" json:"alliance_id,omitempty" toml:"alliance_id" yaml:"alliance_id,omitempty"`
	CorporationID  uint      `db:"corporation_id" boil:"corporation_id" json:"corporation_id" toml:"corporation_id" yaml:"corporation_id"`
	FactionID      null.Uint `db:"faction_id" boil:"faction_id" json:"faction_id,omitempty" toml:"faction_id" yaml:"faction_id,omitempty"`
	AncestryID     uint      `db:"ancestry_id" boil:"ancestry_id" json:"ancestry_id" toml:"ancestry_id" yaml:"ancestry_id"`
	BloodlineID    uint      `db:"bloodline_id" boil:"bloodline_id" json:"bloodline_id" toml:"bloodline_id" yaml:"bloodline_id"`
	RaceID         uint      `db:"race_id" boil:"race_id" json:"race_id" toml:"race_id" yaml:"race_id"`
	Ignored        bool      `db:"ignored" boil:"ignored" json:"ignored" toml:"ignored" yaml:"ignored"`
	Etag           string    `db:"etag" boil:"etag" json:"etag" toml:"etag" yaml:"etag"`
	Expires        time.Time `db:"expires" boil:"expires" json:"expires" toml:"expires" yaml:"expires"`
	CreatedAt      time.Time `db:"created_at" boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`
	UpdatedAt      time.Time `db:"updated_at" boil:"updated_at" json:"updated_at" toml:"updated_at" yaml:"updated_at"`
}

func (c Character) IsExpired() bool {
	return c.Expires.Before(time.Now())
}
