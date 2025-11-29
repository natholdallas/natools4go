// Package fibers is tiny packaging support fiber
package fibers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/natholdallas/natools4go/va"
)

// IdentityParam is simple embedded struct to get id param in your body struct mixin
type IdentityParam struct {
	ID uint `param:"id" json:"-"`
}

// FormData to get any source data as data, sort by [fiber.BodyParser]
func FormData[T any](c *fiber.Ctx) (v T, err error) {
	if err = c.ParamsParser(&v); err != nil {
		return
	}
	if err = c.QueryParser(&v); err != nil {
		return
	}
	if err = c.BodyParser(&v); err != nil {
		return
	}
	err = va.Struct(v)
	return
}

// CookieParser is used to bind cookies to a struct
func CookieParser[T any](c *fiber.Ctx) (v T, err error) {
	err = c.CookieParser(&v)
	return
}

// CookieVarser is used to bind cookies to a struct then verify
func CookieVarser[T any](c *fiber.Ctx) (v T, err error) {
	if err = c.CookieParser(&v); err != nil {
		return
	}
	err = va.Struct(v)
	return
}

// ReqHeaderParser binds the request header strings to a struct.
func ReqHeaderParser[T any](c *fiber.Ctx) (v T, err error) {
	err = c.ReqHeaderParser(&v)
	return
}

// ReqHeaderVarser binds the request header strings to a struct then verify.
func ReqHeaderVarser[T any](c *fiber.Ctx) (v T, err error) {
	if err = c.ReqHeaderParser(&v); err != nil {
		return
	}
	err = va.Struct(v)
	return
}

// BodyParser binds the request body to a struct.
// It supports decoding the following content types based on the Content-Type header:
// application/json, application/xml, application/x-www-form-urlencoded, multipart/form-data
// All JSON extenstion mime types are supported (eg. application/problem+json)
// If none of the content types above are matched, it will return a ErrUnprocessableEntity error
func BodyParser[T any](c *fiber.Ctx) (v T, err error) {
	err = c.BodyParser(&v)
	return
}

// BodyVarser binds the request body to a struct then verify.
// It supports decoding the following content types based on the Content-Type header:
// application/json, application/xml, application/x-www-form-urlencoded, multipart/form-data
// All JSON extenstion mime types are supported (eg. application/problem+json)
// If none of the content types above are matched, it will return a ErrUnprocessableEntity error
func BodyVarser[T any](c *fiber.Ctx) (v T, err error) {
	if err = c.BodyParser(&v); err != nil {
		return
	}
	err = va.Struct(v)
	return
}

// RestParser to get params and body
// be commonly used to [POST, PUT, DELETE, PATCH]
func RestParser[T any](c *fiber.Ctx) (v T, err error) {
	if err = c.ParamsParser(&v); err != nil {
		return
	}
	err = c.BodyParser(&v)
	return
}

// RestVarser to get params and body and verify
// be commonly used to [POST, PUT, DELETE, PATCH]
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
	return
}

// ParamsVarser to get queries and verify
func ParamsVarser[T any](c *fiber.Ctx) (v T, err error) {
	if err = c.ParamsParser(&v); err != nil {
		return
	}
	err = va.Struct(v)
	return
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
