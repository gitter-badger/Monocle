package eveindex

import "time"

type Corporation struct {
	ID            uint      `db:"id" json:"id"`
	Name          string    `db:"name" json:"name"`
	Ticker        string    `db:"ticker" json:"ticker"`
	MemberCount   uint      `db:"member_count" json:"member_count"`
	CeoID         uint      `db:"ceo_id" json:"ceo_id"`
	AllianceID    NullInt64 `db:"alliance_id" json:"alliance_id"`
	DateFounded   NullTime  `db:"date_founded" json:"date_founded"`
	CreatorID     uint64    `db:"creator_id" json:"creator_id"`
	HomeStationID NullInt64 `db:"home_station_id" json:"home_station_id"`
	URL           string    `db:"url" json:"url"`
	TaxRate       float32   `db:"tax_rate" json:"tax_rate"`
	WarEligible   bool      `db:"war_eligible" json:"war_eligible"`
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
