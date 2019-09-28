package monocle

import "time"

type CharacterCorporationHistory struct {
	ID            uint64    `db:"id" json:"id"`
	RecordID      uint      `db:"record_id" json:"record_id"`
	CorporationID uint      `db:"corporation_id" json:"corporation_id"`
	StartDate     time.Time `db:"start_date" json:"start_date"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
}
