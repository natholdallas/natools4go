package fibers

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/natholdallas/natools4go/strs"
)

// ErrorHandler is optimized error handler impl
func ErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusBadRequest
	var e *fiber.Error
	if errors.As(err, &e) {
		if e.Code != 0 {
			code = e.Code
		}
	}
	return c.Status(code).JSON(fiber.Error{Code: code, Message: err.Error()})
}

// Cache used to set response header cache middleware
func Cache(time int64) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set(fiber.HeaderCacheControl, "public, max-age="+strs.FormatInt(time))
		return c.Next()
	}
}

// Logger create log route middleware
func Logger(prefix ...string) fiber.Handler {
	v := ""
	if len(prefix) == 1 {
		v = prefix[0]
	}
	return logger.New(logger.Config{
		TimeFormat: time.DateTime,
		Format:     "[" + v + "]: ${ip}:${port} ${time} ${status} - ${method} ${path} ${error}\n",
	})
}
