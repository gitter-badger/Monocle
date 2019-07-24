package monocle

import "time"

type EtagResource struct {
	ID        uint      `db:"id" json:"id"`
	Resource  string    `db:"resource" json:"resource"`
	Etag      string    `db:"etag" json:"etag"`
	Expires   time.Time `db:"expires" json:"expires"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}
