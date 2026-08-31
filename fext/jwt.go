package fext

import (
	"fmt"
	"time"

	jwtware "github.com/gofiber/contrib/v3/jwt"
	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

// JWTware initializes a Gofiber middleware to protect routes using JWT.
// It uses the standard HS256 signing method and RegisteredClaims.
func JWTware(secretKey string) fiber.Handler {
	return jwtware.New(jwtware.Config{
		SigningKey:   jwtware.SigningKey{Key: []byte(secretKey)},
		ErrorHandler: JWTErrorHandler,
		Claims:       &jwt.RegisteredClaims{},
	})
}

// JWTErrorHandler is a specialized error handler for JWT middleware.
// It captures authentication errors and returns a 401 Unauthorized status.
func JWTErrorHandler(c fiber.Ctx, err error) error {
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
	_, err = jwt.ParseWithClaims(token, &claims, func(t *jwt.Token) (any, error) {
		// Reject non-HMAC signing methods to prevent algorithm confusion attacks.
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secretKey), nil
	})
	return claims, err
}

// JWT represents a managed JWT environment, coupling the secret key with its middleware.
type JWT struct {
	SecretKey  string
	Middleware fiber.Handler
}

// NewJWT initializes a new JWT manager with the given secret key.
func NewJWT(secretKey string) JWT {
	return JWT{secretKey, JWTware(secretKey)}
}

// GenToken is a method wrapper for generating a token using the JWT instance's secret key.
func (j *JWT) GenToken(ID string, endtime ...time.Duration) (string, error) {
	return GenToken(ID, j.SecretKey, endtime...)
}

// ParseToken is a method wrapper for parsing a token using the JWT instance's secret key.
func (j *JWT) ParseToken(token string) (jwt.RegisteredClaims, error) {
	return ParseToken(token, j.SecretKey)
}

// Claims retrieves the JWT RegisteredClaims from the Fiber context.
// It returns nil if the request is not authenticated or the claims are not
// the expected type, so callers can distinguish "no user" from "empty claims".
func (j *JWT) Claims(c fiber.Ctx) *jwt.RegisteredClaims {
	token := jwtware.FromContext(c)
	if token == nil || token.Claims == nil {
		return nil
	}
	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return nil
	}
	return claims
}
