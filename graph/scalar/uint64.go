package scalar

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/99designs/gqlgen/graphql"
)

func MarshalUint64(i uint64) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		json.NewEncoder(w).Encode(i)
	})
}

func UnmarshalUint64(v interface{}) (uint64, error) {
	if _, ok := v.(uint64); !ok {
		return 0, fmt.Errorf("unable to coerce %v to a int64", v)
	}

	return v.(uint64), nil
}
