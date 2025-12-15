package fibers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

// Logger create log route middleware
func Logger(name string, format ...string) fiber.Handler {
	f := "${ip} ${time} ${status} - ${method} ${path} ${error}\n"
	if len(format) > 0 {
		f = format[0]
	}
	return logger.New(logger.Config{
		TimeFormat: time.DateTime,
		Format:     "[" + name + "] " + f,
	})
}
