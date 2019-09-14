package scalar

import (
	"errors"
	"io"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/volatiletech/null"
)

func MarshalNullTime(t null.Time) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		if !t.Valid {
			io.WriteString(w, `null`)
		} else {
			b, _ := t.MarshalJSON()
			io.WriteString(w, string(b))
		}

	})
}

func UnmarshalNullTime(v interface{}) (null.Time, error) {
	if _, ok := v.(string); !ok {
		return null.Time{}, errors.New("time should be in RFC3339 formatted string")
	}

	t, err := time.Parse(time.RFC3339, v.(string))
	if err != nil {
		return null.Time{}, err
	}

	return null.TimeFrom(t), nil
}
