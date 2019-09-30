package monocle

import "time"

// Total is an object representing the database table.
type Total struct {
	ID           uint64    `db:"id" boil:"id" json:"id" toml:"id" yaml:"id"`
	Characters   uint64    `db:"characters" boil:"characters" json:"characters" toml:"characters" yaml:"characters"`
	Corporations uint64    `db:"corporations" boil:"corporations" json:"corporations" toml:"corporations" yaml:"corporations"`
	Alliances    uint64    `db:"alliances" boil:"alliances" json:"alliances" toml:"alliances" yaml:"alliances"`
	CreatedAt    time.Time `db:"created_at" boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`
}
