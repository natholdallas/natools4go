package fibers

import (
	"time"

	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// Jwtware initializes a Gofiber middleware to protect routes using JWT.
// It uses the standard HS256 signing method and RegisteredClaims.
func Jwtware(secretKey string) fiber.Handler {
	return jwtware.New(jwtware.Config{
		SigningKey:   jwtware.SigningKey{Key: []byte(secretKey)},
		ErrorHandler: JwtErrorHandler,
		Claims:       &jwt.RegisteredClaims{},
	})
}

// JwtErrorHandler is a specialized error handler for JWT middleware.
// It captures authentication errors and returns a 401 Unauthorized status.
func JwtErrorHandler(c *fiber.Ctx, err error) error {
	return fiber.NewError(fiber.StatusUnauthorized, err.Error())
}

// GenToken generates a signed JWT string for a specific subject ID.
// The default expiration (if not provided) is 720 hours (30 days).
func GenToken(ID string, secretKey string, endtime ...time.Duration) (string, error) {
	claims := jwt.RegisteredClaims{ID: ID}
	if len(endtime) > 0 {
		claims.ExpiresAt = &jwt.NumericDate{Time: time.Now().Add(endtime[0])}
	} else {
		claims.ExpiresAt = &jwt.NumericDate{Time: time.Now().Add(time.Hour * 720)}
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

// ParseToken decodes and validates a JWT string against the provided secret key.
// It returns the [jwt.RegisteredClaims] if successful.
func ParseToken(token, secretKey string) (claims jwt.RegisteredClaims, err error) {
	_, err = jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (any, error) {
		return []byte(secretKey), nil
	})
	return claims, err
}

// Jwt represents a managed JWT environment, coupling the secret key with its middleware.
type Jwt struct {
	SecretKey  string
	Middleware fiber.Handler
}

// NewJwt initializes a new Jwt manager with the given secret key.
func NewJwt(secretKey string) Jwt {
	return Jwt{secretKey, Jwtware(secretKey)}
}

// GenToken is a method wrapper for generating a token using the Jwt instance's secret key.
func (j *Jwt) GenToken(ID string, endtime ...time.Duration) (string, error) {
	return GenToken(ID, j.SecretKey, endtime...)
}

// ParseToken is a method wrapper for parsing a token using the Jwt instance's secret key.
func (j *Jwt) ParseToken(token string) (jwt.RegisteredClaims, error) {
	return ParseToken(token, j.SecretKey)
}

// Claims retrieves the JWT RegisteredClaims from the Fiber context.
// This should be called on routes protected by the Jwt instance's middleware.
func (j *Jwt) Claims(c *fiber.Ctx) *jwt.RegisteredClaims {
	usr := c.Locals("user")
	if usr == nil {
		return nil
	}
	// Attempt to cast the context local value to the expected jwt.Token
	if token, ok := usr.(*jwt.Token); ok {
		if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok {
			return claims
		}
	}
	return nil
}
