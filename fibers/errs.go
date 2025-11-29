package fibers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
)

// TODO: design error handle logic

// Err to sending fiber original error object
func Err(value any, status ...int) *fiber.Error {
	msg := ""
	code := fiber.StatusBadRequest
	if str, ok := value.(string); ok {
		msg = str
	} else if err, ok := value.(error); ok {
		msg = err.Error()
	}
	if len(status) > 0 {
		code = status[0]
	}
	return &fiber.Error{Code: code, Message: msg}
}

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
