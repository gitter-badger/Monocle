package scalar

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/99designs/gqlgen/graphql"
	"github.com/volatiletech/null"
)

func MarshalNullString(s null.String) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		json.NewEncoder(w).Encode(s.String)
	})
}

func UnmarshalNullString(v interface{}) (null.String, error) {
	if _, ok := v.(string); !ok {
		return null.StringFromPtr(nil), fmt.Errorf("%T is not a string", v)
	}

	return null.StringFrom(v.(string)), nil
}
