// Package vipers is tiny packaging support viper
package vipers

import (
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

type EventHandler func(e fsnotify.Event)

var events []EventHandler = []EventHandler{}

// Watch sets up configuration watching and links it to the internal event dispatcher.
// Call this after setting up your config paths/files.
func Watch() {
	viper.OnConfigChange(Reload)
	viper.WatchConfig()
}

func Reload(e fsnotify.Event) {
	for _, event := range events {
		event(e)
	}
}

func NewUpdateEvent(es ...EventHandler) {
	events = append(events, es...)
}

// --- Helper ---

// set is a private helper to apply a default value if provided.
func set[T any](key string, defaultValue []T) {
	if len(defaultValue) > 0 {
		viper.SetDefault(key, defaultValue[0])
	}
}

// --- Getters ---

// Get uses generics and spf13/cast to retrieve a value of any basic type.
func Get[T cast.Basic](key string, defaultValue ...T) T {
	set(key, defaultValue)
	return cast.To[T](viper.Get(key))
}

// GetString returns the value associated with the key as a string.
func GetString(key string, defaultValue ...string) string {
	set(key, defaultValue)
	return viper.GetString(key)
}

// GetBool returns the value associated with the key as a boolean.
func GetBool(key string, defaultValue ...bool) bool {
	set(key, defaultValue)
	return viper.GetBool(key)
}

// GetInt returns the value associated with the key as an integer.
func GetInt(key string, defaultValue ...int) int {
	set(key, defaultValue)
	return viper.GetInt(key)
}

// GetInt32 returns the value associated with the key as an integer.
func GetInt32(key string, defaultValue ...int32) int32 {
	set(key, defaultValue)
	return viper.GetInt32(key)
}

// GetInt64 returns the value associated with the key as an integer.
func GetInt64(key string, defaultValue ...int64) int64 {
	set(key, defaultValue)
	return viper.GetInt64(key)
}

// GetUint returns the value associated with the key as an unsigned integer.
func GetUint(key string, defaultValue ...uint) uint {
	set(key, defaultValue)
	return viper.GetUint(key)
}

// GetUint8 returns the value associated with the key as an unsigned integer.
func GetUint8(key string, defaultValue ...uint8) uint8 {
	set(key, defaultValue)
	return viper.GetUint8(key)
}

// GetUint16 returns the value associated with the key as an unsigned integer.
func GetUint16(key string, defaultValue ...uint16) uint16 {
	set(key, defaultValue)
	return viper.GetUint16(key)
}

// GetUint32 returns the value associated with the key as an unsigned integer.
func GetUint32(key string, defaultValue ...uint32) uint32 {
	set(key, defaultValue)
	return viper.GetUint32(key)
}

// GetUint64 returns the value associated with the key as an unsigned integer.
func GetUint64(key string, defaultValue ...uint64) uint64 {
	set(key, defaultValue)
	return viper.GetUint64(key)
}

// GetFloat64 returns the value associated with the key as a float64.
func GetFloat64(key string, defaultValue ...float64) float64 {
	set(key, defaultValue)
	return viper.GetFloat64(key)
}

// GetTime returns the value associated with the key as time.
func GetTime(key string, defaultValue ...time.Time) time.Time {
	set(key, defaultValue)
	return viper.GetTime(key)
}

// GetDuration returns the value associated with the key as a duration.
func GetDuration(key string, defaultValue ...time.Duration) time.Duration {
	set(key, defaultValue)
	return viper.GetDuration(key)
}

// GetIntSlice returns the value associated with the key as a slice of int values.
func GetIntSlice(key string, defaultValue ...[]int) []int {
	set(key, defaultValue)
	return viper.GetIntSlice(key)
}

// GetStringSlice returns the value associated with the key as a slice of strings.
func GetStringSlice(key string, defaultValue ...[]string) []string {
	set(key, defaultValue)
	return viper.GetStringSlice(key)
}

// GetStringMap returns the value associated with the key as a map of interfaces.
func GetStringMap(key string, defaultValue ...map[string]any) map[string]any {
	set(key, defaultValue)
	return viper.GetStringMap(key)
}

// GetStringMapString returns the value associated with the key as a map of strings.
func GetStringMapString(key string, defaultValue ...map[string]string) map[string]string {
	set(key, defaultValue)
	return viper.GetStringMapString(key)
}

// GetStringMapStringSlice returns the value associated with the key as a map to a slice of strings.
func GetStringMapStringSlice(key string, defaultValue ...map[string][]string) map[string][]string {
	set(key, defaultValue)
	return viper.GetStringMapStringSlice(key)
}

// GetSizeInBytes returns the size of the value associated with the given key
// in bytes.
func GetSizeInBytes(key string, defaultValue ...uint) uint {
	set(key, defaultValue)
	return viper.GetSizeInBytes(key)
}
