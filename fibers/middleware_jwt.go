package fibers

import (
	"time"

	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// JwtBasic for next version design to use middleware
// it can used by any user table or others data struct [Claims.ID] type, now only use uint
// try add genericity in jwt middleware template design in the future
type JwtBasic interface {
	string | bool | uint | uint8 | uint16 | uint32 | uint64 | int | int8 | int16 | int32 | int64
}

type Claims struct {
	jwt.RegisteredClaims
	ID uint `json:"jti,omitempty"`
}

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
func GenToken(ID uint, secretKey string, endtime ...time.Duration) (string, error) {
	claims := Claims{ID: ID, RegisteredClaims: jwt.RegisteredClaims{}}
	if len(endtime) > 0 {
		claims.ExpiresAt = &jwt.NumericDate{Time: time.Now().Add(endtime[0])}
	} else {
		claims.ExpiresAt = &jwt.NumericDate{Time: time.Now().Add(time.Hour * 720)}
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

// ParseToken can parse to [Claims] by token and secretKey
func ParseToken(token, secretKey string) (Claims, error) {
	claims := Claims{}
	_, err := jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (any, error) {
		return []byte(secretKey), nil
	})
	return claims, err
}

// GetToken ignore [ParseToken]'s error
func GetToken(token string, secretKey string) Claims {
	claims, _ := ParseToken(token, secretKey)
	return claims
}
