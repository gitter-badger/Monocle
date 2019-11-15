package monocle

import (
	"time"

	"github.com/volatiletech/null"
)

type Alliance struct {
	ID                    uint      `db:"id" json:"id"`
	Name                  string    `db:"name" json:"name"`
	Ticker                string    `db:"ticker" json:"ticker"`
	CreatorCorporationID  uint      `db:"creator_corporation_id" json:"creator_corporation_id"`
	CreatorID             uint64    `db:"creator_id" json:"creator_id"`
	DateFounded           null.Time `db:"date_founded" json:"date_founded,omitempty"`
	ExecutorCorporationID uint      `db:"executor_corporation_id" json:"executor_corporation_id"`
	MemberCount           uint      `db:"member_count" json:"member_count"`
	Ignored               bool      `db:"ignored" json:"ignored"`
	Closed                bool      `db:"closed" json:"closed"`
	Etag                  string    `db:"etag" json:"etag"`
	Expires               time.Time `db:"expires" json:"expires"`
	CreatedAt             time.Time `db:"created_at" json:"created_at"`
	UpdatedAt             time.Time `db:"updated_at" json:"updated_at"`
}

func (a Alliance) IsExpired() bool {
	return a.Expires.Before(time.Now())
}
