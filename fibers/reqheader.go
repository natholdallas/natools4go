package fibers

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/natholdallas/natools4go/arrs"
	"github.com/natholdallas/natools4go/strs"
)

// GetAuthorization extracts the credential part from the "Authorization" header.
// It removes the scheme prefix (e.g., "Bearer ") from the header value.
// If no custom scheme is provided, it defaults to "Bearer ".
func GetAuthorization(c *fiber.Ctx, scheme ...string) string {
	auth := c.Get(fiber.HeaderAuthorization)
	if auth == "" {
		return ""
	}
	prefix := strs.ToEnd(arrs.GetDefault("Bearer ", scheme), strs.Space)
	// Case-insensitive prefix removal for better compatibility
	if strings.HasPrefix(strings.ToLower(auth), strings.ToLower(prefix)) {
		return auth[len(prefix):]
	}
	return auth
}
