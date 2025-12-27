package orms

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// List is a generic slice type that supports JSON serialization/deserialization
// for database persistence.
type List[T any] []T

// Scan implements the sql.Scanner interface to decode a JSON-encoded value
// from the database into the List.
func (s *List[T]) Scan(value any) error {
	var data []byte
	switch v := value.(type) {
	case nil:
		*s = nil
		return nil
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return fmt.Errorf("cannot convert %T to []byte", value)
	}
	return json.Unmarshal(data, s)
}

// Value implements the driver.Valuer interface to encode the List into
// a JSON-encoded format for database storage.
func (l List[T]) Value() (driver.Value, error) {
	// Optional: You could return nil if len(l) == 0 to save space as NULL
	return json.Marshal(l)
}

// Dict is a generic map type that supports Dict serialization/deserialization
// for database persistence.
type Dict[T any] map[string]T

// Scan implements the sql.Scanner interface to decode a JSON-encoded value
// from the database into the JSON map.
func (d *Dict[T]) Scan(value any) error {
	var bytes []byte
	switch v := value.(type) {
	case nil:
		*d = nil
		return nil
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("failed to unmarshal JSONB value: expected []byte or string, got %T", value)
	}

	// If the database stores an empty string, treat it as an empty/nil map
	if len(bytes) == 0 {
		*d = nil
		return nil
	}

	return json.Unmarshal(bytes, d)
}

// Value implements the driver.Valuer interface to encode the JSON map into
// a JSON-encoded format for database storage.
func (d Dict[T]) Value() (driver.Value, error) {
	return json.Marshal(d)
}
