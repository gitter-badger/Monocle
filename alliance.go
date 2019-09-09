package monocle

import (
	"time"

	"github.com/volatiletech/null"
)

type Alliance struct {
	ID                    uint64    `db:"id" json:"id"`
	Name                  string    `db:"name" json:"name"`
	Ticker                string    `db:"ticker" json:"ticker"`
	CreatorCorporationID  uint64    `db:"creator_corporation_id" json:"creator_corporation_id"`
	CreatorID             uint64    `db:"creator_id" json:"creator_id"`
	DateFounded           null.Time `db:"date_founded" json:"date_founded"`
	ExecutorCorporationID uint64    `db:"executor_corporation_id" json:"executor_corporation_id"`
	Ignored               bool      `db:"ignored" json:"ignored"`
	Closed                bool      `db:"closed" json:"closed"`
	Expires               time.Time `db:"expires" json:"expires"`
	Etag                  string    `db:"etag" json:"etag"`
	CreatedAt             time.Time `db:"created_at" json:"created_at"`
	UpdatedAt             time.Time `db:"updated_at" json:"updated_at"`
	Exists                bool      `json:"-"`
}

func (a Alliance) IsExpired() bool {
	if a.Expires.Before(time.Now()) {
		return true
	}
	return false
}
