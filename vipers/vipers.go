// Package vipers is tiny packaging support viper
package vipers

import (
	"errors"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/natholdallas/natools4go/slice"
	"github.com/natholdallas/natools4go/va"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

type EventHandler func(e fsnotify.Event)

// eventsMu guards events so Reload (fired from viper's watcher goroutine) can
// snapshot handlers under the lock and invoke them outside it. This avoids
// holding the lock while running user callbacks, which could otherwise deadlock
// when a handler calls RegisterUpdateHandler.
var (
	eventsMu sync.RWMutex
	events   []EventHandler = []EventHandler{}
)

// Watch sets up configuration watching and links it to the internal event dispatcher.
// Call this after setting up your config paths/files.
func Watch(es ...EventHandler) {
	RegisterUpdateHandler(es...)
	viper.OnConfigChange(Reload)
	viper.WatchConfig()
}

func Config(name, path string, confType ...string) error {
	viper.SetConfigName(name)
	viper.AddConfigPath(path)
	viper.SetConfigType(slice.Defu("toml", confType))
	return viper.ReadInConfig()
}

func MustConfig(name, path string, confType ...string) {
	err := Config(name, path, confType...)
	if err != nil {
		panic(errors.New("error reading config: " + err.Error()))
	}
}

func Validate(data any) error {
	return va.Struct(data)
}

func MustValidate(data any) {
	err := Validate(data)
	if err != nil {
		panic(errors.New("missing required configuration:\n" + err.Error()))
	}
}

func Reload(e fsnotify.Event) {
	eventsMu.RLock()
	handlers := make([]EventHandler, len(events))
	copy(handlers, events)
	eventsMu.RUnlock()
	for _, event := range handlers {
		event(e)
	}
}

// RegisterUpdateHandler registers one or more event handlers. Handlers are
// invoked in registration order on every config change; duplicate registrations
// are not deduplicated (EventHandler is a function type, which cannot be compared).
func RegisterUpdateHandler(es ...EventHandler) {
	eventsMu.Lock()
	defer eventsMu.Unlock()
	for _, h := range es {
		if h == nil {
			continue
		}
		events = append(events, h)
	}
}

// --- Helper ---

// defu returns the default value when a non-empty default is provided and the
// key is not present in any configuration source (config file, env, flags).
// Unlike viper.SetDefault it does not write permanently to the global config,
// avoiding per-call locking and side effects.
func defu[T any](key string, defaultValue []T) (T, bool) {
	if len(defaultValue) > 0 && !viper.IsSet(key) {
		return defaultValue[0], true
	}
	var zero T
	return zero, false
}

// --- Getters ---

// Get uses generics and spf13/cast to retrieve a value of any basic type.
func Get[T cast.Basic](key string, defaultValue ...T) T {
	if v, ok := defu(key, defaultValue); ok {
		return v
	}
	return cast.To[T](viper.Get(key))
}

// String returns the value associated with the key as a string.
func String(key string, defaultValue ...string) string {
	if v, ok := defu(key, defaultValue); ok {
		return v
	}
	return viper.GetString(key)
}

// Bool returns the value associated with the key as a boolean.
func Bool(key string, defaultValue ...bool) bool {
	if v, ok := defu(key, defaultValue); ok {
		return v
	}
	return viper.GetBool(key)
}

// Int returns the value associated with the key as an integer.
func Int(key string, defaultValue ...int) int {
	if v, ok := defu(key, defaultValue); ok {
		return v
	}
	return viper.GetInt(key)
}

// Int32 returns the value associated with the key as an integer.
func Int32(key string, defaultValue ...int32) int32 {
	if v, ok := defu(key, defaultValue); ok {
		return v
	}
	return viper.GetInt32(key)
}

// Int64 returns the value associated with the key as an integer.
func Int64(key string, defaultValue ...int64) int64 {
	if v, ok := defu(key, defaultValue); ok {
		return v
	}
	return viper.GetInt64(key)
}

// Uint returns the value associated with the key as an unsigned integer.
func Uint(key string, defaultValue ...uint) uint {
	if v, ok := defu(key, defaultValue); ok {
		return v
	}
	return viper.GetUint(key)
}

// Uint8 returns the value associated with the key as an unsigned integer.
func Uint8(key string, defaultValue ...uint8) uint8 {
	if v, ok := defu(key, defaultValue); ok {
		return v
	}
	return viper.GetUint8(key)
}

// Uint16 returns the value associated with the key as an unsigned integer.
func Uint16(key string, defaultValue ...uint16) uint16 {
	if v, ok := defu(key, defaultValue); ok {
		return v
	}
	return viper.GetUint16(key)
}

// Uint32 returns the value associated with the key as an unsigned integer.
func Uint32(key string, defaultValue ...uint32) uint32 {
	if v, ok := defu(key, defaultValue); ok {
		return v
	}
	return viper.GetUint32(key)
}

// Uint64 returns the value associated with the key as an unsigned integer.
func Uint64(key string, defaultValue ...uint64) uint64 {
	if v, ok := defu(key, defaultValue); ok {
		return v
	}
	return viper.GetUint64(key)
}

// Float64 returns the value associated with the key as a float64.
func Float64(key string, defaultValue ...float64) float64 {
	if v, ok := defu(key, defaultValue); ok {
		return v
	}
	return viper.GetFloat64(key)
}

// Time returns the value associated with the key as time.
func Time(key string, defaultValue ...time.Time) time.Time {
	if v, ok := defu(key, defaultValue); ok {
		return v
	}
	return viper.GetTime(key)
}

// Duration returns the value associated with the key as a duration.
func Duration(key string, defaultValue ...time.Duration) time.Duration {
	if v, ok := defu(key, defaultValue); ok {
		return v
	}
	return viper.GetDuration(key)
}

// IntSlice returns the value associated with the key as a slice of int values.
func IntSlice(key string, defaultValue ...[]int) []int {
	if v, ok := defu(key, defaultValue); ok {
		return v
	}
	return viper.GetIntSlice(key)
}

// StringSlice returns the value associated with the key as a slice of strings.
func StringSlice(key string, defaultValue ...[]string) []string {
	if v, ok := defu(key, defaultValue); ok {
		return v
	}
	return viper.GetStringSlice(key)
}

// StringMap returns the value associated with the key as a map of interfaces.
func StringMap(key string, defaultValue ...map[string]any) map[string]any {
	if v, ok := defu(key, defaultValue); ok {
		return v
	}
	return viper.GetStringMap(key)
}

// StringMapString returns the value associated with the key as a map of strings.
func StringMapString(key string, defaultValue ...map[string]string) map[string]string {
	if v, ok := defu(key, defaultValue); ok {
		return v
	}
	return viper.GetStringMapString(key)
}

// StringMapStringSlice returns the value associated with the key as a map to a slice of strings.
func StringMapStringSlice(key string, defaultValue ...map[string][]string) map[string][]string {
	if v, ok := defu(key, defaultValue); ok {
		return v
	}
	return viper.GetStringMapStringSlice(key)
}

// SizeInBytes returns the size of the value associated with the given key
// in bytes.
func SizeInBytes(key string, defaultValue ...uint) uint {
	if v, ok := defu(key, defaultValue); ok {
		return v
	}
	return viper.GetSizeInBytes(key)
}
