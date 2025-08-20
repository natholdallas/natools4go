// Package fibers is tiny packaging support fiber
package fibers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/natholdallas/natools4go/va"
)

// IdentityParam is simple embedded struct to get id param in your body struct mixin
type IdentityParam struct {
	ID uint `param:"id" json:"-"`
}

// BodyParser to get body
func BodyParser[T any](c *fiber.Ctx) (v T, err error) {
	err = c.BodyParser(&v)
	return
}

// BodyVarser to get body and verify
func BodyVarser[T any](c *fiber.Ctx) (v T, err error) {
	if err = c.BodyParser(&v); err != nil {
		return
	}
	err = va.Struct(v)
	return
}

// RestParser to get params and body, be commonly used to [POST, PUT, DELETE, PATCH]
func RestParser[T any](c *fiber.Ctx) (v T, err error) {
	if err = c.ParamsParser(&v); err != nil {
		return
	}
	err = c.BodyParser(&v)
	return
}

// RestVarser to get params and body and verify, be commonly used to [POST, PUT, DELETE, PATCH]
func RestVarser[T any](c *fiber.Ctx) (v T, err error) {
	if err = c.ParamsParser(&v); err != nil {
		return
	}
	if err = c.BodyParser(&v); err != nil {
		return
	}
	err = va.Struct(v)
	return
}

// QueryParser to get queries, be commonly used to [GET]
func QueryParser[T any](c *fiber.Ctx) (v T, err error) {
	err = c.QueryParser(&v)
	return
}

// QueryVarser to get queries and verify, be commonly used to [GET]
func QueryVarser[T any](c *fiber.Ctx) (v T, err error) {
	if err = c.QueryParser(&v); err != nil {
		return
	}
	err = va.Struct(v)
	return
}

// ParamsParser to get queries, be commonly used to [GET]
func ParamsParser[T any](c *fiber.Ctx) (v T, err error) {
	err = c.QueryParser(&v)
	return v, err
}

// ParamsVarser to get queries and verify
func ParamsVarser[T any](c *fiber.Ctx) (v T, err error) {
	if err = c.ParamsParser(&v); err != nil {
		return
	}
	err = va.Struct(v)
	return
}

// ParamsUint get params as uint
func ParamsUint(c *fiber.Ctx, key string, defaultValue ...int) (uint, error) {
	value, err := strconv.ParseUint(c.Params(key), 10, 64)
	return uint(value), err
}

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
