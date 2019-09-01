package monocle

import "time"

type CorporationAllianceHistory struct {
	ID         uint64    `db:"id" json:"id"`
	RecordID   uint      `db:"record_id" json:"record_id"`
	AllianceID NullInt64 `db:"alliance_id" json:"alliance_id"`
	StartDate  time.Time `db:"start_date" json:"start_date"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
	Exists     bool
}
