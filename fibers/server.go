// Package fibers is tiny packaging support fiber
package fibers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/natholdallas/natools4go/gorms"
	"github.com/natholdallas/natools4go/va"
)

// BodyParser to get body
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

// BodyParserVa to get body and verify
func BodyParserVa[T any](c *fiber.Ctx) (T, error) {
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

// RestParser to get params and body, be commonly used to [POST, PUT, DELETE, PATCH]
func RestParser[T any](c *fiber.Ctx) (T, error) {
	var result T
	if err := c.ParamsParser(&result); err != nil {
		return result, err
	}
	if err := c.BodyParser(&result); err != nil {
		return result, err
	}
	return result, nil
}

// RestParserVa to get params and body and verify, be commonly used to [POST, PUT, DELETE, PATCH]
func RestParserVa[T any](c *fiber.Ctx) (T, error) {
	var result T
	if err := c.ParamsParser(&result); err != nil {
		return result, err
	}
	if err := c.BodyParser(&result); err != nil {
		return result, err
	}
	if err := va.Struct(result); err != nil {
		return result, err
	}
	return result, nil
}

// QueryParser to get queries, be commonly used to [GET]
func QueryParser[T any](c *fiber.Ctx) (T, error) {
	var queries T
	err := c.QueryParser(&queries)
	return queries, err
}

// QueryParserVa to get queries and verify, be commonly used to [GET]
func QueryParserVa[T any](c *fiber.Ctx) (T, error) {
	var queries T
	if err := c.QueryParser(&queries); err != nil {
		return queries, err
	}
	if err := va.Struct(queries); err != nil {
		return queries, err
	}
	return queries, nil
}

// Pagination to getting the [gorms.Pagination] struct
func Pagination(c *fiber.Ctx) gorms.Pagination {
	page := max(c.QueryInt("page", 1), 1)
	size := min(c.QueryInt("size", 20), 100)
	if size < 0 {
		size = 20
	}
	return gorms.Pagination{Page: page, Size: size}
}

// ParamsUint get params as uint
func ParamsUint(c *fiber.Ctx, key string, defaultValue ...int) (uint, error) {
	value, err := strconv.ParseUint(c.Params(key), 10, 64)
	return uint(value), err
}
