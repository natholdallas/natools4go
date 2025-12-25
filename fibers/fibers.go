// Package fibers is tiny packaging support fiber
package fibers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/natholdallas/natools4go/va"
)

// StdLogFmt defines a standard format string for Fiber's logger middleware.
const StdLogFmt = "${ip} ${time} ${status} - ${method} ${path} ${error}\n"

// IdentityParam is a mixin struct for embedding common ID parameters from URIs.
// Usage: type UserReq struct { fibers.IdentityParam; Name string `json:"name"` }
type IdentityParam struct {
	ID uint `param:"id" json:"-"`
}

// FormData binds data from all possible sources: URI parameters, Query strings, and Request Body.
// It prioritizes BodyParser as the final override.
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

// --- Cookie Parsers ---

// CookieParser binds request cookies to the provided struct type T.
func CookieParser[T any](c *fiber.Ctx) (v T, err error) {
	err = c.CookieParser(&v)
	return
}

// CookieVarser binds request cookies to struct T and performs validation.
func CookieVarser[T any](c *fiber.Ctx) (v T, err error) {
	if err = c.CookieParser(&v); err != nil {
		return
	}
	err = va.Struct(v)
	return
}

// --- Header Parsers ---

// ReqHeaderParser binds request headers to the provided struct type T.
func ReqHeaderParser[T any](c *fiber.Ctx) (v T, err error) {
	err = c.ReqHeaderParser(&v)
	return
}

// ReqHeaderVarser binds request headers to struct T and performs validation.
func ReqHeaderVarser[T any](c *fiber.Ctx) (v T, err error) {
	if err = c.ReqHeaderParser(&v); err != nil {
		return
	}
	err = va.Struct(v)
	return
}

// BodyParser binds the request body (JSON, XML, Form, etc.) to struct T.
// It supports decoding the following content types based on the Content-Type header:
// application/json, application/xml, application/x-www-form-urlencoded, multipart/form-data
// All JSON extenstion mime types are supported (eg. application/problem+json)
// If none of the content types above are matched, it will return a ErrUnprocessableEntity error
func BodyParser[T any](c *fiber.Ctx) (v T, err error) {
	err = c.BodyParser(&v)
	return
}

// BodyVarser binds the request body to struct T and performs validation.
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

// --- RESTful Parsers (Params + Body) ---

// RestParser binds both URI parameters and the request body to struct T.
// Ideal for POST, PUT, and PATCH requests.
func RestParser[T any](c *fiber.Ctx) (v T, err error) {
	if err = c.ParamsParser(&v); err != nil {
		return
	}
	err = c.BodyParser(&v)
	return
}

// RestVarser binds URI parameters and the request body to struct T, then performs validation.
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

// --- Query Parsers ---

// QueryParser binds URL query parameters to struct T. Usually used for GET requests.
func QueryParser[T any](c *fiber.Ctx) (v T, err error) {
	err = c.QueryParser(&v)
	return
}

// QueryVarser binds URL query parameters to struct T and performs validation.
func QueryVarser[T any](c *fiber.Ctx) (v T, err error) {
	if err = c.QueryParser(&v); err != nil {
		return
	}
	err = va.Struct(v)
	return
}

// --- URI Param Parsers ---

// ParamsParser binds URI route parameters to struct T.
func ParamsParser[T any](c *fiber.Ctx) (v T, err error) {
	err = c.ParamsParser(&v)
	return
}

// ParamsVarser binds URI route parameters to struct T and performs validation.
func ParamsVarser[T any](c *fiber.Ctx) (v T, err error) {
	if err = c.ParamsParser(&v); err != nil {
		return
	}
	err = va.Struct(v)
	return
}

// --- Response Helpers ---

// Status is a shorthand to set the HTTP response status code.
func Status(c *fiber.Ctx, status int) error {
	c.Status(status)
	return nil
}

// JSON sends a JSON response with the specified HTTP status code.
func JSON(c *fiber.Ctx, status int, data any) error {
	return c.Status(status).JSON(data)
}

// SendString sends a plain text response with the specified HTTP status code.
func SendString(c *fiber.Ctx, status int, str string) error {
	return c.Status(status).SendString(str)
}
