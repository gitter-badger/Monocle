package monocle

type Entity struct {
	ID       int    `db:"id" json:"id"`
	Name     string `db:"name" json:"name"`
	Category string `db:"category" json:"category"`
}
