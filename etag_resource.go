package monocle

import "time"

type EtagResource struct {
	ID        uint64    `db:"id" json:"id"`
	Resource  string    `db:"resource" json:"resource"`
	Etag      string    `db:"etag" json:"etag"`
	Expires   time.Time `db:"expires" json:"expires"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
	Exists    bool
}

func (e EtagResource) IsExpired() bool {
	if e.Expires.Before(time.Now()) {
		return true
	}
	return false
}
