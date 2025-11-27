package fibers

import (
	"time"

	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// Jwtware is simply middleware func
// example: var IsLogin = Jwtware("xxx")
func Jwtware(secretKey string) func(c *fiber.Ctx) error {
	return jwtware.New(jwtware.Config{
		SigningKey:   jwtware.SigningKey{Key: []byte(secretKey)},
		ErrorHandler: JwtErrorHandler,
	})
}

// JwtErrorHandler if you only used error handler
func JwtErrorHandler(c *fiber.Ctx, err error) error {
	return Err(err, fiber.StatusUnauthorized)
}

// GenToken can generate token by secretKey
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

// ParseToken can parse to [jwt.RegisteredClaims] by token and secretKey
func ParseToken(token, secretKey string) (claims jwt.RegisteredClaims, err error) {
	_, err = jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (any, error) {
		return []byte(secretKey), nil
	})
	return claims, err
}
