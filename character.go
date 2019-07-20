package eveindex

import "time"

type Character struct {
	ID             uint64    `db:"id" json:"id"`
	Name           string    `db:"name" json:"name"`
	Birthday       time.Time `db:"birthday" json:"birthday"`
	Gender         string    `db:"gender" json:"gender"`
	SecurityStatus float32   `db:"security_status" json:"security_status"`
	AllianceID     NullInt64 `db:"alliance_id" json:"alliance_id"`
	CorporationID  uint32    `db:"corporation_id" json:"corporation_id"`
	FactionID      uint32    `db:"faction_id" json:"faction_id"`
	AncestryID     uint32    `db:"ancestry_id" json:"ancestry_id"`
	BloodlineID    uint32    `db:"bloodline_id" json:"bloodline_id"`
	RaceID         uint32    `db:"race_id" json:"race_id"`
	Expires        time.Time `db:"expires" json:"expires"`
	Etag           string    `db:"etag" json:"etag"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time `db:"updated_at" json:"updated_at"`
}

func (c Character) IsExpired() bool {
	if c.Expires.Before(time.Now()) {
		return true
	}
	return false
}
