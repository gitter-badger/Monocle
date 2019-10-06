package monocle

// UnknownCorp is an object representing the database table.
type UnknownCorp struct {
	ID uint64 `db:"id" boil:"id" json:"id" toml:"id" yaml:"id"`
}
