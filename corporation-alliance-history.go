package monocle

import (
	"time"

	"github.com/volatiletech/null"
)

type CorporationAllianceHistory struct {
	ID         uint64      `db:"id" json:"id"`
	RecordID   uint        `db:"record_id" json:"record_id"`
	AllianceID null.Uint32 `db:"alliance_id" json:"alliance_id"`
	StartDate  time.Time   `db:"start_date" json:"start_date"`
	LeaveDate  null.Time   `db:"leave_date" json:"leave_date"`
	CreatedAt  time.Time   `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time   `db:"updated_at" json:"updated_at"`
}
