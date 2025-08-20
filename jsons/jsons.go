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

func Set(source map[string]any, value any, keys ...string) {
	src := source
	for i, key := range keys {
		// if last key
		if len(keys)-1 == i {
			src[key] = value
			break
		}
		src = Get(src, key)
	}
}

func Get(source map[string]any, keys ...string) map[string]any {
	src := source
	for _, key := range keys {
		src = src[key].(map[string]any)
	}
	return src
}
