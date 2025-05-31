package fibers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
)

// an optimized error handler impl
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
