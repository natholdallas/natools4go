package fext

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"github.com/natholdallas/natools4go/constraints"
)

var (
	debug   bool            = false
	errFunc func(err error) = nil
)

func SetDebugMode(mode bool) {
	debug = mode
}

func SetErrorFunc(fn func(err error)) {
	errFunc = fn
}

func SetLogLevel[T constraints.Integer](lv T) {
	log.SetLevel(log.Level(lv))
}

// Listen starts the server on the given address.
func Listen(app *fiber.App, addr string, config ...fiber.ListenConfig) {
	if err := app.Listen(addr); err != nil {
		errFunc(err)
	}
}

// FormatPort takes an integer port and returns a formatted string like ":xxxx".
func FmtPort(port int) string {
	return ":" + strconv.Itoa(port)
}
