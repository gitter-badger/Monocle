package scalar

import (
	"database/sql"
	"fmt"
	"io"
)

type NullInt64 sql.NullInt64

func (n *NullInt64) MarshalGQL(w io.Writer) {
	fmt.Fprintf(w, "%d", n.Int64)
}

func (n *NullInt64) UnmarshalGQL(v interface{}) error {
	if i, ok := v.(int64); ok {
		n.Int64 = i
		n.Valid = true
		return nil
	}

	return nil
}

type NullInt64 sql.NullInt64

sql.Null

func (n *NullInt64) MarshalGQL(w io.Writer) {
	fmt.Fprintf(w, "%d", n.Int64)
}

func (n *NullInt64) UnmarshalGQL(v interface{}) error {
	if i, ok := v.(int64); ok {
		n.Int64 = i
		n.Valid = true
		return nil
	}

	return nil
}
