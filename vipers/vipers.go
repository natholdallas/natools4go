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

func Reload(e fsnotify.Event) {
	for _, event := range events {
		event(e)
	}
}

func NewUpdateEvent(event EventHandler) {
	events = append(events, event)
}

func Get[T cast.Basic](key string, defaultValue ...T) T {
	if len(defaultValue) > 0 {
		viper.SetDefault(key, defaultValue[0])
	}
	return cast.To[T](viper.Get(key))
}

func GetString(key string, defaultValue ...string) string {
	if len(defaultValue) > 0 {
		viper.SetDefault(key, defaultValue[0])
	}
	return viper.GetString(key)
}

func GetBool(key string, defaultValue ...bool) bool {
	if len(defaultValue) > 0 {
		viper.SetDefault(key, defaultValue[0])
	}
	return viper.GetBool(key)
}

func GetInt(key string, defaultValue ...int) int {
	if len(defaultValue) > 0 {
		viper.SetDefault(key, defaultValue[0])
	}
	return viper.GetInt(key)
}

func GetInt32(key string, defaultValue ...int32) int32 {
	if len(defaultValue) > 0 {
		viper.SetDefault(key, defaultValue[0])
	}
	return viper.GetInt32(key)
}

func GetInt64(key string, defaultValue ...int64) int64 {
	if len(defaultValue) > 0 {
		viper.SetDefault(key, defaultValue[0])
	}
	return viper.GetInt64(key)
}

func GetUint(key string, defaultValue ...uint) uint {
	if len(defaultValue) > 0 {
		viper.SetDefault(key, defaultValue[0])
	}
	return viper.GetUint(key)
}

func GetUint8(key string, defaultValue ...uint8) uint8 {
	if len(defaultValue) > 0 {
		viper.SetDefault(key, defaultValue[0])
	}
	return viper.GetUint8(key)
}

func GetUint16(key string, defaultValue ...uint16) uint16 {
	if len(defaultValue) > 0 {
		viper.SetDefault(key, defaultValue[0])
	}
	return viper.GetUint16(key)
}

func GetUint32(key string, defaultValue ...uint32) uint32 {
	if len(defaultValue) > 0 {
		viper.SetDefault(key, defaultValue[0])
	}
	return viper.GetUint32(key)
}

func GetUint64(key string, defaultValue ...uint64) uint64 {
	if len(defaultValue) > 0 {
		viper.SetDefault(key, defaultValue[0])
	}
	return viper.GetUint64(key)
}

func GetTime(key string, defaultValue ...time.Time) time.Time {
	if len(defaultValue) > 0 {
		viper.SetDefault(key, defaultValue[0])
	}
	return viper.GetTime(key)
}

func GetDuration(key string, defaultValue ...time.Duration) time.Duration {
	if len(defaultValue) > 0 {
		viper.SetDefault(key, defaultValue[0])
	}
	return viper.GetDuration(key)
}

func GetStringSlice(key string, defaultValue ...[]string) []string {
	if len(defaultValue) > 0 {
		viper.SetDefault(key, defaultValue[0])
	}
	return viper.GetStringSlice(key)
}

func GetStringMap(key string, defaultValue ...map[string]any) map[string]any {
	if len(defaultValue) > 0 {
		viper.SetDefault(key, defaultValue[0])
	}
	return viper.GetStringMap(key)
}

func GetStringMapString(key string, defaultValue ...map[string]string) map[string]string {
	if len(defaultValue) > 0 {
		viper.SetDefault(key, defaultValue[0])
	}
	return viper.GetStringMapString(key)
}

func GetStringMapStringSlice(key string, defaultValue ...map[string][]string) map[string][]string {
	if len(defaultValue) > 0 {
		viper.SetDefault(key, defaultValue[0])
	}
	return viper.GetStringMapStringSlice(key)
}
