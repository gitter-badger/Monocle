package scalar

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/99designs/gqlgen/graphql"
)

func MarshalFloat32(i float32) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		json.NewEncoder(w).Encode(i)
	})
}

func UnmarshalFloat32(v interface{}) (float32, error) {
	if _, ok := v.(float32); !ok {
		return 0, fmt.Errorf("unable to coerce %v to a float32", v)
	}

	return v.(float32), nil
}

func MarshalFloat64(i float64) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		json.NewEncoder(w).Encode(i)
	})
}

func UnmarshalFloat64(v interface{}) (float64, error) {
	if _, ok := v.(float64); !ok {
		return 0, fmt.Errorf("unable to coerce %v to a float64", v)
	}

	return v.(float64), nil
}
