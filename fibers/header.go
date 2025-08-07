package fibers

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

func GetAuthorization(c *fiber.Ctx, scheme ...string) string {
	s := c.Get(fiber.HeaderAuthorization)
	prefix := "Bearer "
	if len(scheme) > 0 {
		prefix = scheme[0]
	}
	return strings.TrimPrefix(s, prefix)
}
