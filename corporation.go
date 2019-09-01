package monocle

import "time"

type Corporation struct {
	ID            uint64    `db:"id" json:"id"`
	Name          string    `db:"name" json:"name"`
	Ticker        string    `db:"ticker" json:"ticker"`
	MemberCount   uint64    `db:"member_count" json:"member_count"`
	CeoID         uint64    `db:"ceo_id" json:"ceo_id"`
	AllianceID    NullInt64 `db:"alliance_id" json:"alliance_id"`
	DateFounded   NullTime  `db:"date_founded" json:"date_founded"`
	CreatorID     uint64    `db:"creator_id" json:"creator_id"`
	HomeStationID NullInt64 `db:"home_station_id" json:"home_station_id"`
	TaxRate       float32   `db:"tax_rate" json:"tax_rate"`
	WarEligible   bool      `db:"war_eligible" json:"war_eligible"`
	Ignored       bool      `db:"ignored" json:"ignored"`
	Closed        bool      `db:"closed" json:"closed"`
	Expires       time.Time `db:"expires" json:"expires"`
	Etag          string    `db:"etag" json:"etag"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
}

func (c Corporation) IsExpired() bool {
	if c.Expires.Before(time.Now()) {
		return true
	}
	return false
}
