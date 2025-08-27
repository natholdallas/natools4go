package gorms

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type List[T any] []T

func (list *List[T]) Scan(value any) error {
	if value == nil {
		*list = nil
		return nil
	}
	v, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot convert %T to []byte", value)
	}
	return json.Unmarshal(v, list)
}

func (list List[T]) Value() (driver.Value, error) {
	return json.Marshal(list)
}
