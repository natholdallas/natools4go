package fext

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
)

var (
	DebugMode bool            = false
	ErrorFunc func(err error) = nil
)

// Listen starts the server on the given address.
func Listen(app *fiber.App, addr string, config ...fiber.ListenConfig) {
	if err := app.Listen(addr); err != nil {
		ErrorFunc(err)
	}
}

// FormatPort takes an integer port and returns a formatted string like ":xxxx".
func FmtPort(port int) string {
	return ":" + strconv.Itoa(port)
}
