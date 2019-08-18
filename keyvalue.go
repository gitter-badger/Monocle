package monocle

import (
	"encoding/json"
	"time"
)

type KeyValue struct {
	Key       string          `db:"k" json:"k"`
	Value     json.RawMessage `db:"v" json:"v"`
	CreatedAt time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt time.Time       `db:"updated_at" json:"updated_at"`
	Exists    bool
}
