// Package fext is tiny packaging support fiber
package fext

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/natholdallas/natools4go/slice"
	"github.com/natholdallas/natools4go/strs"
	"github.com/natholdallas/natools4go/va"
)

// StdLogFmt defines a standard format string for Fiber's logger middleware.
const StdLogFmt = "[${ip}:${port}] ${time} ${status} - ${method} ${path} ${error}\n"

// IdentityParam is a mixin struct for embedding common ID parameters from URIs.
type IdentityParam[T any] struct {
	ID T `param:"id" json:"-" mapstructure:"-"`
} // @name IdentityParam

// Value is simple struct for get field from value
type Value[T any] struct {
	Value T `json:"value" mapstructure:"value"`
} // @name Value

// FormData binds data from all possible sources: URI parameters, Query strings, and Request Body.
// It prioritizes BodyParser as the final override.
func FormData[T any](c fiber.Ctx) (v T, err error) {
	if err = c.Bind().URI(&v); err != nil {
		return
	}
	if err = c.Bind().Query(&v); err != nil {
		return
	}
	if err = c.Bind().Body(&v); err != nil {
		return
	}
	err = va.Struct(v)
	return
}

// --- Cookie Parsers ---

// CookieParser binds request cookies to the provided struct type T.
func CookieParser[T any](c fiber.Ctx) (v T, err error) {
	err = c.Bind().Cookie(&v)
	return
}

// CookieVarser binds request cookies to struct T and performs validation.
func CookieVarser[T any](c fiber.Ctx) (v T, err error) {
	if err = c.Bind().Cookie(&v); err != nil {
		return
	}
	err = va.Struct(v)
	return
}

// --- Header Parsers ---

// ReqHeaderParser binds request headers to the provided struct type T.
func ReqHeaderParser[T any](c fiber.Ctx) (v T, err error) {
	err = c.Bind().Header(&v)
	return
}

// ReqHeaderVarser binds request headers to struct T and performs validation.
func ReqHeaderVarser[T any](c fiber.Ctx) (v T, err error) {
	if err = c.Bind().Header(&v); err != nil {
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
func BodyParser[T any](c fiber.Ctx) (v T, err error) {
	err = c.Bind().Body(&v)
	return
}

// BodyVarser binds the request body to struct T and performs validation.
// It supports decoding the following content types based on the Content-Type header:
// application/json, application/xml, application/x-www-form-urlencoded, multipart/form-data
// All JSON extenstion mime types are supported (eg. application/problem+json)
// If none of the content types above are matched, it will return a ErrUnprocessableEntity error
func BodyVarser[T any](c fiber.Ctx) (v T, err error) {
	if err = c.Bind().Body(&v); err != nil {
		return
	}
	err = va.Struct(v)
	return
}

// --- RESTful Parsers (Params + Body) ---

// RestParser binds both URI parameters and the request body to struct T.
// Ideal for POST, PUT, and PATCH requests.
func RestParser[T any](c fiber.Ctx) (v T, err error) {
	if err = c.Bind().URI(&v); err != nil {
		return
	}
	err = c.Bind().Body(&v)
	return
}

// RestVarser binds URI parameters and the request body to struct T, then performs validation.
func RestVarser[T any](c fiber.Ctx) (v T, err error) {
	if err = c.Bind().URI(&v); err != nil {
		return
	}
	if err = c.Bind().Body(&v); err != nil {
		return
	}
	err = va.Struct(v)
	return
}

// --- Query Parsers ---

// QueryParser binds URL query parameters to struct T. Usually used for GET requests.
func QueryParser[T any](c fiber.Ctx) (v T, err error) {
	err = c.Bind().Query(&v)
	return
}

// QueryVarser binds URL query parameters to struct T and performs validation.
func QueryVarser[T any](c fiber.Ctx) (v T, err error) {
	if err = c.Bind().Query(&v); err != nil {
		return
	}
	err = va.Struct(v)
	return
}

// --- URI Param Parsers ---

// ParamsParser binds URI route parameters to struct T.
func ParamsParser[T any](c fiber.Ctx) (v T, err error) {
	err = c.Bind().URI(&v)
	return
}

// ParamsVarser binds URI route parameters to struct T and performs validation.
func ParamsVarser[T any](c fiber.Ctx) (v T, err error) {
	if err = c.Bind().URI(&v); err != nil {
		return
	}
	err = va.Struct(v)
	return
}

// --- Response Helpers ---

// Status is a shorthand to set the HTTP response status code.
func Status(c fiber.Ctx, status int) error {
	c.Status(status)
	return nil
}

// JSON sends a JSON response with the specified HTTP status code.
func JSON(c fiber.Ctx, status int, data any) error {
	return c.Status(status).JSON(data)
}

// SendString sends a plain text response with the specified HTTP status code.
func SendString(c fiber.Ctx, status int, str string) error {
	return c.Status(status).SendString(str)
}

// --- Request Helpers ---

// GetAuthorization extracts the credential part from the "Authorization" header.
// It removes the scheme prefix (e.g., "Bearer ") from the header value.
// If no custom scheme is provided, it defaults to "Bearer ".
func GetAuthorization(c fiber.Ctx, scheme ...string) string {
	auth := c.Get(fiber.HeaderAuthorization)
	if auth == "" {
		return ""
	}
	prefix := strs.ToEnd(slice.Defu("Bearer ", scheme), strs.Space)
	// Case-insensitive prefix removal for better compatibility
	if strings.HasPrefix(strings.ToLower(auth), strings.ToLower(prefix)) {
		return auth[len(prefix):]
	}
	return auth
}
