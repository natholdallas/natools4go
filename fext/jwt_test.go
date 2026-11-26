package fext

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

func TestJWTwareEmptySecretPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("JWTware(\"\") should panic")
		}
	}()
	JWTware("")
}

func jwtApp(secret string) *fiber.App {
	app := fiber.New()
	app.Use(JWTware(secret))
	app.Get("/", func(c fiber.Ctx) error { return c.SendString("ok") })
	return app
}

func TestJWTwareRejectsWrongAlgorithm(t *testing.T) {
	const secret = "s3cret"
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.RegisteredClaims{
		Subject:   "u1",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	})
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign HS512: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(fiber.HeaderAuthorization, "Bearer "+signed)
	resp, err := jwtApp(secret).Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("status = %d, want 401", resp.StatusCode)
	}
}

func TestGenTokenClaims(t *testing.T) {
	signed, err := GenToken("u-42", "s3cret")
	if err != nil {
		t.Fatalf("GenToken: %v", err)
	}
	claims, err := ParseToken(signed, "s3cret")
	if err != nil {
		t.Fatalf("ParseToken: %v", err)
	}
	if claims.Subject != "u-42" {
		t.Errorf("Subject = %q, want u-42", claims.Subject)
	}
	if claims.ID != "u-42" {
		t.Errorf("ID = %q, want u-42", claims.ID)
	}
	if claims.IssuedAt == nil {
		t.Error("IssuedAt should be set")
	}
	if claims.ExpiresAt == nil {
		t.Error("ExpiresAt should be set")
	}
}

func TestParseTokenRequiresExpiration(t *testing.T) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{Subject: "u1"})
	signed, err := token.SignedString([]byte("s3cret"))
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	if _, err := ParseToken(signed, "s3cret"); err == nil {
		t.Error("ParseToken accepted a token without exp")
	}
}
