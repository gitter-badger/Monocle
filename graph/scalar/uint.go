package scalar

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/99designs/gqlgen/graphql"
	"github.com/volatiletech/null"
)

func MarshalUint(i uint) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		json.NewEncoder(w).Encode(i)
	})
}

func UnmarshalUint(v interface{}) (uint, error) {
	if _, ok := v.(uint); !ok {
		return 0, fmt.Errorf("unable to coerce %v to a int", v)
	}

	return v.(uint), nil
}

func MarshalNullUint(n null.Uint) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		json.NewEncoder(w).Encode(n.Uint)
	})
}

func UnmarshalNullUint(v interface{}) (null.Uint, error) {
	if _, ok := v.(uint); !ok {
		return null.UintFromPtr(nil), fmt.Errorf("%T is not a uint", v)
	}

	return null.UintFrom(v.(uint)), nil
}

func MarshalUint32(i uint32) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		json.NewEncoder(w).Encode(i)
	})
}

func UnmarshalUint32(v interface{}) (uint32, error) {
	if _, ok := v.(uint32); !ok {
		return 0, fmt.Errorf("unable to coerce %v to a int32", v)
	}

	return v.(uint32), nil
}

func MarshalNullUint32(n null.Uint32) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		json.NewEncoder(w).Encode(n.Uint32)
	})
}

func UnmarshalNullUint32(v interface{}) (null.Uint32, error) {
	if _, ok := v.(uint32); !ok {
		return null.Uint32FromPtr(nil), fmt.Errorf("%T is not a uint32", v)
	}

	return null.Uint32From(v.(uint32)), nil
}

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
