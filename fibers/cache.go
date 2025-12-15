package fibers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/natholdallas/natools4go/strs"
)

// Cache used to set response header cache middleware
func Cache(time int64) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set(fiber.HeaderCacheControl, "public, max-age="+strs.FormatInt(time))
		return c.Next()
	}
}
