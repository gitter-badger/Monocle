package monocle

import "time"

type Counter struct {
	ID        uint64    `db:"id" json:"id"`
	CharCount uint      `db:"char_count" json:"char_count"`
	CorpCount uint      `db:"corp_count" json:"corp_count"`
	AlliCount uint      `db:"alli_count" json:"alli_count"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}
