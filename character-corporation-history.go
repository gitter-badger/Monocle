package monocle

import (
	"time"

	"github.com/volatiletech/null"
)

type CharacterCorporationHistory struct {
	ID            uint64    `db:"id" boil:"id" json:"id" toml:"id" yaml:"id"`
	RecordID      uint      `db:"record_id" boil:"record_id" json:"record_id" toml:"record_id" yaml:"record_id"`
	CorporationID uint      `db:"corporation_id" boil:"corporation_id" json:"corporation_id" toml:"corporation_id" yaml:"corporation_id"`
	StartDate     time.Time `db:"start_date" boil:"start_date" json:"start_date" toml:"start_date" yaml:"start_date"`
	LeaveDate     null.Time `db:"leave_date" boil:"leave_date" json:"leave_date,omitempty" toml:"leave_date" yaml:"leave_date,omitempty"`
	CreatedAt     time.Time `db:"created_at" boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" boil:"updated_at" json:"updated_at" toml:"updated_at" yaml:"updated_at"`
}
