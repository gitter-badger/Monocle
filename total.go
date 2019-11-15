package monocle

import "time"

// Total is an object representing the database table.
type Total struct {
	ID           uint64    `db:"id" json:"id"`
	Characters   uint64    `db:"characters" json:"characters"`
	Corporations uint64    `db:"corporations" json:"corporations"`
	Alliances    uint64    `db:"alliances" json:"alliances"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}
