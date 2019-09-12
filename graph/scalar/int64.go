package scalar

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/99designs/gqlgen/graphql"
)

func MarshalInt64(i int64) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		json.NewEncoder(w).Encode(i)
	})
}

func UnmarshalInt64(v interface{}) (int64, error) {
	if _, ok := v.(int64); !ok {
		return 0, fmt.Errorf("unable to coerce %v to a int64", v)
	}

	return v.(int64), nil
}
