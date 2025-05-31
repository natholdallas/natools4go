package fibers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/natholdallas/natools4go/gorms"
	"github.com/natholdallas/natools4go/va"
)

// get body and verify
func BodyParser[T any](c *fiber.Ctx) (T, error) {
	var v T
	err := c.BodyParser(&v)
	if err != nil {
		return v, err
	}
	err = va.Struct(v)
	if err != nil {
		return v, err
	}
	return v, err
}

// get params & body and verify, be commonly used to [POST, PUT, DELETE, PATCH]
func RestParser[T any](c *fiber.Ctx) (T, error) {
	var v T
	if err := c.ParamsParser(&v); err != nil {
		return v, err
	}
	if err := c.BodyParser(&v); err != nil {
		return v, err
	}
	if err := va.Struct(v); err != nil {
		return v, err
	}
	return v, nil
}

// get queries and verify, be commonly used to [GET]
func QueryParser[T any](c *fiber.Ctx) (T, error) {
	var v T
	if err := c.QueryParser(&v); err != nil {
		return v, err
	}
	if err := va.Struct(v); err != nil {
		return v, err
	}
	return v, nil
}

// get gorm pagination
func Pagination(c *fiber.Ctx) gorms.Pagination {
	return gorms.Pagination{
		Page: c.QueryInt("page", 1),
		Size: c.QueryInt("size", 20),
	}
}

// get params as uint
func ParamsUint(c *fiber.Ctx, key string, defaultValue ...int) (uint, error) {
	value, err := strconv.ParseUint(c.Params(key), 10, 64)
	return uint(value), err
}

// send status only use one line
func Status(c *fiber.Ctx, status int) error {
	c.Status(status)
	return nil
}

// send json and status
func JSON(c *fiber.Ctx, status int, data any) error {
	c.Status(status)
	return c.JSON(data)
}

// send string and status
func SendString(c *fiber.Ctx, status int, str string) error {
	c.Status(status)
	return c.SendString(str)
}

// send an error
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
