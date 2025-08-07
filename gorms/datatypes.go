package gorms

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type List[T any] []T

func (sl *List[T]) Scan(value any) error {
	if value == nil {
		*sl = nil
		return nil
	}
	s, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot convert %T to []byte", value)
	}
	return json.Unmarshal(s, sl)
}

func (sl List[T]) Value() (driver.Value, error) {
	return json.Marshal(sl)
}
