package monocle

import "time"

type CorporationDelta struct {
	ID            uint64    `db:"id" json:"id"`
	CorporationID uint64    `db:"corporation_id" json:"corporation_id"`
	MemberCount   uint64    `db:"member_count" json:"member_count"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
}
