// Package jsons is tiny packaging support json
package jsons

import (
	"encoding/json"
)

func Unmarshal[T any](bytes []byte) (T, error) {
	var result T
	err := json.Unmarshal(bytes, &result)
	return result, err
}

func Marshal(v any, pretty ...bool) ([]byte, error) {
	if len(pretty) > 0 {
		if pretty[0] {
			return json.MarshalIndent(v, "", "\t")
		}
	}
	return json.Marshal(v)
}

func String(data any, pretty ...bool) (string, error) {
	if len(pretty) > 0 {
		if pretty[0] {
			d, err := json.MarshalIndent(data, "", "  ")
			return string(d), err
		}
	}
	d, err := json.Marshal(data)
	return string(d), err
}
