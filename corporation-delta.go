package monocle

import "time"

type CorporationDelta struct {
	ID            uint64    `db:"id" boil:"id" json:"id" toml:"id" yaml:"id"`
	CorporationID uint64    `db:"corporation_id" boil:"corporation_id" json:"corporation_id" toml:"corporation_id" yaml:"corporation_id"`
	MemberCount   uint64    `db:"member_count" boil:"member_count" json:"member_count" toml:"member_count" yaml:"member_count"`
	CreatedAt     time.Time `db:"created_at" boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`
}
