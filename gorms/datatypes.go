package gorms

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type List[T any] []T

func (s *List[T]) Scan(value any) error {
	if value == nil {
		*s = nil
		return nil
	}
	v, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot convert %T to []byte", value)
	}
	return json.Unmarshal(v, s)
}

func (l List[T]) Value() (driver.Value, error) {
	return json.Marshal(l)
}

type JSON[T any] map[string]T

func (s *JSON[T]) Scan(value any) error {
	if value == nil {
		*s = nil
		return nil
	}
	v, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot convert %T to []byte", value)
	}
	return json.Unmarshal(v, s)
}

func (s JSON[T]) Value() (driver.Value, error) {
	return json.Marshal(s)
}
