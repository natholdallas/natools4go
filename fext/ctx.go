package fext

import (
	"io"
	"strconv"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/natholdallas/natools4go/constraints"
)

var (
	mu      sync.RWMutex
	debug   bool            = false
	errFunc func(err error) = nil
)

func SetDebugMode(mode bool) {
	mu.Lock()
	debug = mode
	mu.Unlock()
}

func isDebug() bool {
	mu.RLock()
	defer mu.RUnlock()
	return debug
}

func SetErrorFunc(fn func(err error)) {
	mu.Lock()
	errFunc = fn
	mu.Unlock()
}

func onError(err error) {
	mu.RLock()
	fn := errFunc
	mu.RUnlock()
	if fn != nil {
		fn(err)
	}
}

func SetLogLevel[T constraints.Integer](lv T) {
	log.SetLevel(log.Level(lv))
}

// Listen starts the server on the given address.
func Listen(app *fiber.App, addr string, config ...fiber.ListenConfig) {
	if err := app.Listen(addr, config...); err != nil {
		onError(err)
	}
}

// Log is a middleware that logs the request and response.
func Log(prefix string, stream io.Writer) fiber.Handler {
	return logger.New(logger.Config{
		TimeFormat: time.DateTime,
		Format:     "[" + prefix + "]" + StdLogFmt,
		Stream:     stream,
	})
}

// FormatPort takes an integer port and returns a formatted string like ":xxxx".
func FmtPort(port int) string {
	return ":" + strconv.Itoa(port)
}
