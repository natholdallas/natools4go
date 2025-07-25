package fibers

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

var logfunc = logger.New(logger.Config{Format: "[${ip}:${port}] ${time} ${status} - ${method} ${path}\n"})

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

func Cache(time int64) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		c.Set(fiber.HeaderCacheControl, "public, max-age="+strconv.FormatInt(time, 10))
		return c.Next()
	}
}

func NoCache(c *fiber.Ctx) error {
	c.Set(fiber.HeaderCacheControl, "no-cache")
	return c.Next()
}

func Logger(c *fiber.Ctx) error {
	return logfunc(c)
}
