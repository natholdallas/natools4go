package fibers

import "github.com/gofiber/fiber/v2"

// Status only use one line
func Status(c *fiber.Ctx, status int) error {
	c.Status(status)
	return nil
}

// JSON to sending json body and status
func JSON(c *fiber.Ctx, status int, data any) error {
	c.Status(status)
	return c.JSON(data)
}

// SendString to sending string and status
func SendString(c *fiber.Ctx, status int, str string) error {
	c.Status(status)
	return c.SendString(str)
}

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
