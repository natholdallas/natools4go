package fext

import (
	"github.com/gofiber/fiber/v3"
)

// Listen starts the server on the given address and returns any startup or
// runtime failure.
func Listen(app *fiber.App, addr string, config ...fiber.ListenConfig) error {
	return app.Listen(addr, config...)
}

// MustListen starts the server on the given address, panicking on failure.
func MustListen(app *fiber.App, addr string, config ...fiber.ListenConfig) {
	if err := Listen(app, addr, config...); err != nil {
		panic(err)
	}
}
