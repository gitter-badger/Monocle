package eveindex

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/go-sql-driver/mysql"
)

type NullInt64 struct {
	sql.NullInt64
}

func (v NullInt64) MarshalJSON() ([]byte, error) {
	if v.Valid {
		return json.Marshal(v.Int64)
	} else {
		return json.Marshal(nil)
	}
}

func (v *NullInt64) UnmarshalJSON(data []byte) error {

	// Unmarshalling into a pointer will let us detect null
	var x *int64

	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	if x != nil {
		v.Valid = true
		v.Int64 = *x
	} else {
		v.Valid = false
	}
	return nil
}

type NullFloat64 struct {
	sql.NullFloat64
}

func (v NullFloat64) MarshalJSON() ([]byte, error) {
	if v.Valid {
		return json.Marshal(v.Float64)
	} else {
		return json.Marshal(nil)
	}
}

func (v *NullFloat64) UnmarshalJSON(data []byte) error {

	// Unmarshalling into a pointer will let us detect null
	var x *float64

	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	if x != nil {
		v.Valid = true
		v.Float64 = *x
	} else {
		v.Valid = false
	}
	return nil
}

type NullString struct {
	sql.NullString
}

func (v NullString) MarshalJSON() ([]byte, error) {
	if v.Valid {
		return json.Marshal(v.String)
	} else {
		return json.Marshal(nil)
	}
}

func (v *NullString) UnmarshalJSON(data []byte) error {

	// Unmarshalling into a pointer will let us detect null
	var x *string

	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	if x != nil {
		v.Valid = true
		v.String = *x
	} else {
		v.Valid = false
	}
	return nil
}

type NullBool struct {
	sql.NullBool
}

func (v NullBool) MarshalJSON() ([]byte, error) {
	if v.Valid {
		return json.Marshal(v.Bool)
	} else {
		return json.Marshal(nil)
	}
}

func (v *NullBool) UnmarshalJSON(data []byte) error {

	// Unmarshalling into a pointer will let us detect null
	var x *bool

	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	if x != nil {
		v.Valid = true
		v.Bool = *x
	} else {
		v.Valid = false
	}
	return nil
}

type NullTime struct {
	mysql.NullTime
}

func (v NullTime) MarshalJSON() ([]byte, error) {
	if v.Valid {
		return json.Marshal(v.Time)
	} else {
		return json.Marshal(nil)
	}
}

func (v *NullTime) UnmarshalJSON(data []byte) error {

	// Unmarshalling into a pointer will let us detect null
	var x *time.Time

	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	if x != nil {
		v.Valid = true
		v.Time = *x
	} else {
		v.Valid = false
	}
	return nil
}

type JSONArrayToString struct {
	sql.NullString
}

func (j *JSONArrayToString) MarshalJSON() ([]byte, error) {
	if j.Valid {
		return json.Marshal(j.String)
	}
	return json.Marshal(nil)
}

func (j *JSONArrayToString) UnmarshalJSON(data []byte) error {
	// Unmarshalling into a pointer will let us detect null
	var x *json.RawMessage

	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	if x != nil {
		j.Valid = true
		j.String = string(*x)
	} else {
		j.Valid = false
	}
	return nil
}
