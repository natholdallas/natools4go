// Package jsons is tiny packaging support json
package jsons

import (
	"encoding/json"

	"github.com/natholdallas/natools4go/arrs"
)

// Unmarshal parses the JSON-encoded data and stores the result in a new value of type T.
// It leverages Go generics to return the specific type directly along with any error.
func Unmarshal[T any](bytes []byte) (T, error) {
	var v T
	err := json.Unmarshal(bytes, &v)
	return v, err
}

func IUnmarshal[T any](bytes []byte) T {
	var v T
	json.Unmarshal(bytes, &v)
	return v
}

// Marshal returns the JSON encoding of v.
// If the optional pretty parameter is set to true, it returns indented JSON using tabs.
func Marshal(v any, pretty ...bool) ([]byte, error) {
	p := arrs.GetDefault(false, pretty)
	if p {
		return json.MarshalIndent(v, "", "\t")
	}
	return json.Marshal(v)
}

func IMarshal(v any, pretty ...bool) []byte {
	p := arrs.GetDefault(false, pretty)
	if p {
		value, _ := json.MarshalIndent(v, "", "\t")
		return value
	}
	value, _ := json.Marshal(v)
	return value
}

// String returns the JSON encoding of data as a string.
// If the optional pretty parameter is true, it returns indented JSON using two spaces.
func String(data any, pretty ...bool) (string, error) {
	p := arrs.GetDefault(false, pretty)
	if p {
		d, err := json.MarshalIndent(data, "", "  ")
		return string(d), err
	}
	d, err := json.Marshal(data)
	return string(d), err
}

func IString(data any, pretty ...bool) string {
	p := arrs.GetDefault(false, pretty)
	if p {
		d, _ := json.MarshalIndent(data, "", "  ")
		return string(d)
	}
	d, _ := json.Marshal(data)
	return string(d)
}

// Set traverses a nested map structure using the provided keys and assigns the value to the final key.
// Note: This function assumes all intermediate levels are already existing maps.
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

// Get traverses a nested map structure by key sequence and returns the map at the final key.
// Warning: This function uses a direct type assertion which will cause a panic if a key
// does not exist or if the value is not a map[string]any.
func Get(source map[string]any, keys ...string) map[string]any {
	src := source
	for _, key := range keys {
		src = src[key].(map[string]any)
	}
	return src
}
