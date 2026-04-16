package fext

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
)

// Cache returns a fiber.Handler (middleware) that sets the "Cache-Control" header.
// The 'seconds' parameter determines how long the response should be cached by
// browsers and intermediate proxies.
//
// Example:
//
//	app.Get("/static", fext.Cache(3600), handleStatic) // Cache for 1 hour
func Cache(seconds int64) fiber.Handler {
	// Construct the header value once to improve middleware performance
	cacheValue := fmt.Sprintf("public, max-age=%d", seconds)

	return func(c fiber.Ctx) error {
		c.Set(fiber.HeaderCacheControl, cacheValue)
		return c.Next()
	}
}
