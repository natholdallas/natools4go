// Package vipers is tiny packaging support viper
package vipers

import (
	"log"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/natholdallas/natools4go/slice"
	"github.com/natholdallas/natools4go/spew"
	"github.com/natholdallas/natools4go/va"
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

func Config(name, path string, confType ...string) {
	viper.SetConfigName(name)
	viper.AddConfigPath(path)
	viper.SetConfigType(slice.Defu("toml", confType))
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("error reading config: %v", err)
	}
}

func Validate(data any) {
	if err := va.Struct(data); err != nil {
		log.Fatalf("missing required configuration:\n%v", err)
	} else {
		spew.JSON(data)
	}
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

// defu is a private helper to apply a default value if provided.
func defu[T any](key string, defaultValue []T) {
	if len(defaultValue) > 0 {
		viper.SetDefault(key, defaultValue[0])
	}
}

// --- Getters ---

// Get uses generics and spf13/cast to retrieve a value of any basic type.
func Get[T cast.Basic](key string, defaultValue ...T) T {
	defu(key, defaultValue)
	return cast.To[T](viper.Get(key))
}

// String returns the value associated with the key as a string.
func String(key string, defaultValue ...string) string {
	defu(key, defaultValue)
	return viper.GetString(key)
}

// Bool returns the value associated with the key as a boolean.
func Bool(key string, defaultValue ...bool) bool {
	defu(key, defaultValue)
	return viper.GetBool(key)
}

// Int returns the value associated with the key as an integer.
func Int(key string, defaultValue ...int) int {
	defu(key, defaultValue)
	return viper.GetInt(key)
}

// Int32 returns the value associated with the key as an integer.
func Int32(key string, defaultValue ...int32) int32 {
	defu(key, defaultValue)
	return viper.GetInt32(key)
}

// Int64 returns the value associated with the key as an integer.
func Int64(key string, defaultValue ...int64) int64 {
	defu(key, defaultValue)
	return viper.GetInt64(key)
}

// Uint returns the value associated with the key as an unsigned integer.
func Uint(key string, defaultValue ...uint) uint {
	defu(key, defaultValue)
	return viper.GetUint(key)
}

// Uint8 returns the value associated with the key as an unsigned integer.
func Uint8(key string, defaultValue ...uint8) uint8 {
	defu(key, defaultValue)
	return viper.GetUint8(key)
}

// Uint16 returns the value associated with the key as an unsigned integer.
func Uint16(key string, defaultValue ...uint16) uint16 {
	defu(key, defaultValue)
	return viper.GetUint16(key)
}

// Uint32 returns the value associated with the key as an unsigned integer.
func Uint32(key string, defaultValue ...uint32) uint32 {
	defu(key, defaultValue)
	return viper.GetUint32(key)
}

// Uint64 returns the value associated with the key as an unsigned integer.
func Uint64(key string, defaultValue ...uint64) uint64 {
	defu(key, defaultValue)
	return viper.GetUint64(key)
}

// Float64 returns the value associated with the key as a float64.
func Float64(key string, defaultValue ...float64) float64 {
	defu(key, defaultValue)
	return viper.GetFloat64(key)
}

// Time returns the value associated with the key as time.
func Time(key string, defaultValue ...time.Time) time.Time {
	defu(key, defaultValue)
	return viper.GetTime(key)
}

// Duration returns the value associated with the key as a duration.
func Duration(key string, defaultValue ...time.Duration) time.Duration {
	defu(key, defaultValue)
	return viper.GetDuration(key)
}

// IntSlice returns the value associated with the key as a slice of int values.
func IntSlice(key string, defaultValue ...[]int) []int {
	defu(key, defaultValue)
	return viper.GetIntSlice(key)
}

// StringSlice returns the value associated with the key as a slice of strings.
func StringSlice(key string, defaultValue ...[]string) []string {
	defu(key, defaultValue)
	return viper.GetStringSlice(key)
}

// StringMap returns the value associated with the key as a map of interfaces.
func StringMap(key string, defaultValue ...map[string]any) map[string]any {
	defu(key, defaultValue)
	return viper.GetStringMap(key)
}

// StringMapString returns the value associated with the key as a map of strings.
func StringMapString(key string, defaultValue ...map[string]string) map[string]string {
	defu(key, defaultValue)
	return viper.GetStringMapString(key)
}

// StringMapStringSlice returns the value associated with the key as a map to a slice of strings.
func StringMapStringSlice(key string, defaultValue ...map[string][]string) map[string][]string {
	defu(key, defaultValue)
	return viper.GetStringMapStringSlice(key)
}

// SizeInBytes returns the size of the value associated with the given key
// in bytes.
func SizeInBytes(key string, defaultValue ...uint) uint {
	defu(key, defaultValue)
	return viper.GetSizeInBytes(key)
}
