package orms

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"
)

// List is a generic slice type that supports JSON serialization/deserialization
// for database persistence.
type List[T any] []T // @name List

// Scan implements the sql.Scanner interface to decode a JSON-encoded value
// from the database into the List.
func (s *List[T]) Scan(value any) error {
	return JSONScan(s, value)
}

// Value implements the driver.Valuer interface to encode the List into
// a JSON-encoded format for database storage.
func (s List[T]) Value() (driver.Value, error) {
	return JSONValue(s)
}

// Dict is a generic map type that supports Dict serialization/deserialization
// for database persistence.
type Dict[T any] map[string]T // @name Dict

// Scan implements the sql.Scanner interface to decode a JSON-encoded value
// from the database into the JSON map.
func (d *Dict[T]) Scan(value any) error {
	return JSONScan(d, value)
}

// Value implements the driver.Valuer interface to encode the JSON map into
// a JSON-encoded format for database storage.
func (d Dict[T]) Value() (driver.Value, error) {
	return JSONValue(d)
}

func JSONScan(d any, value any) error {
	rv := reflect.ValueOf(d)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return fmt.Errorf("failed to unmarshal JSONB value: expected non-nil pointer, got %T", d)
	}

	var bytes []byte
	switch v := value.(type) {
	case nil:
		rv.Elem().SetZero()
		return nil
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("failed to unmarshal JSONB value: expected []byte or string, got %T", value)
	}

	// If the database stores an empty value, treat it as empty/nil
	if len(bytes) == 0 {
		rv.Elem().SetZero()
		return nil
	}
	return json.Unmarshal(bytes, d)
}

func JSONValue(d any) (driver.Value, error) {
	if d == nil {
		return nil, nil
	}
	rv := reflect.ValueOf(d)
	if rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return nil, nil
		}
		d = rv.Elem().Interface()
	}
	return json.Marshal(d)
}
