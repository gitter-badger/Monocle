package monocle

import (
	"time"

	"github.com/volatiletech/null"
)

type Alliance struct {
	ID                    uint      `db:"id" boil:"id" json:"id" toml:"id" yaml:"id"`
	Name                  string    `db:"name" boil:"name" json:"name" toml:"name" yaml:"name"`
	Ticker                string    `db:"ticker" boil:"ticker" json:"ticker" toml:"ticker" yaml:"ticker"`
	CreatorCorporationID  uint      `db:"creator_corporation_id" boil:"creator_corporation_id" json:"creator_corporation_id" toml:"creator_corporation_id" yaml:"creator_corporation_id"`
	CreatorID             uint64    `db:"creator_id" boil:"creator_id" json:"creator_id" toml:"creator_id" yaml:"creator_id"`
	DateFounded           null.Time `db:"date_founded" boil:"date_founded" json:"date_founded,omitempty" toml:"date_founded" yaml:"date_founded,omitempty"`
	ExecutorCorporationID uint      `db:"executor_corporation_id" boil:"executor_corporation_id" json:"executor_corporation_id" toml:"executor_corporation_id" yaml:"executor_corporation_id"`
	MemberCount           uint      `db:"member_count" boil:"member_count" json:"member_count" toml:"member_count" yaml:"member_count"`
	Ignored               bool      `db:"ignored" boil:"ignored" json:"ignored" toml:"ignored" yaml:"ignored"`
	Closed                bool      `db:"closed" boil:"closed" json:"closed" toml:"closed" yaml:"closed"`
	Etag                  string    `db:"etag" boil:"etag" json:"etag" toml:"etag" yaml:"etag"`
	Expires               time.Time `db:"expires" boil:"expires" json:"expires" toml:"expires" yaml:"expires"`
	CreatedAt             time.Time `db:"created_at" boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`
	UpdatedAt             time.Time `db:"updated_at" boil:"updated_at" json:"updated_at" toml:"updated_at" yaml:"updated_at"`
}

func (a Alliance) IsExpired() bool {
	return a.Expires.Before(time.Now())
}
