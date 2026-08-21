package fext

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
)

func newTestApp() *fiber.App {
	return fiber.New(fiber.Config{ErrorHandler: ErrorHandler})
}

func TestErrorHandlerFail(t *testing.T) {
	app := newTestApp()
	app.Get("/", func(c fiber.Ctx) error {
		return &Fail{Status: 404, Code: "NOT_FOUND", Message: "nope"}
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != 404 {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
	body := readBody(resp)
	if !strings.Contains(body, `"code":"NOT_FOUND"`) || !strings.Contains(body, `"message":"nope"`) {
		t.Errorf("body = %s, want code/message", body)
	}
}

func TestErrorHandlerFiberError(t *testing.T) {
	app := newTestApp()
	app.Get("/", func(c fiber.Ctx) error {
		return fiber.NewError(fiber.StatusTeapot, "teapot")
	})

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/", nil))
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != fiber.StatusTeapot {
		t.Errorf("status = %d, want %d", resp.StatusCode, fiber.StatusTeapot)
	}
}

func TestErrorHandlerGeneric(t *testing.T) {
	app := newTestApp()
	app.Get("/", func(c fiber.Ctx) error {
		return &plainErr{"boom"}
	})

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/", nil))
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
	if body := readBody(resp); !strings.Contains(body, "boom") {
		t.Errorf("body = %s, want boom", body)
	}
}

type plainErr struct{ msg string }

func (e *plainErr) Error() string { return e.msg }

func TestErrorHandlerSystemHiddenInProd(t *testing.T) {
	SetDebugMode(false)
	defer SetDebugMode(false)
	app := newTestApp()
	app.Get("/", func(c fiber.Ctx) error {
		return &Fail{Status: 500, Message: "internal", System: &plainErr{"secret-detail"}}
	})

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/", nil))
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if body := readBody(resp); strings.Contains(body, "secret-detail") {
		t.Errorf("system detail leaked in prod: %s", body)
	}
}

func TestErrorHandlerSystemShownInDebug(t *testing.T) {
	SetDebugMode(true)
	defer SetDebugMode(true)
	app := newTestApp()
	app.Get("/", func(c fiber.Ctx) error {
		return &Fail{Status: 500, Message: "internal", System: &plainErr{"secret-detail"}}
	})

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/", nil))
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if body := readBody(resp); !strings.Contains(body, "secret-detail") {
		t.Errorf("system detail missing in debug: %s", body)
	}
}

func TestGetAuthorization(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString(GetAuthorization(c))
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(fiber.HeaderAuthorization, "Bearer abc123")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if body := readBody(resp); body != "abc123" {
		t.Errorf("authorization = %q, want abc123", body)
	}
}

func TestGetAuthorizationNoScheme(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString(GetAuthorization(c, "Token "))
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(fiber.HeaderAuthorization, "Token xyz")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if body := readBody(resp); body != "xyz" {
		t.Errorf("authorization = %q, want xyz", body)
	}
}

func TestFmtPort(t *testing.T) {
	if got := FmtPort(8080); got != ":8080" {
		t.Errorf("FmtPort = %q, want :8080", got)
	}
}

func TestJSONAndStatus(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c fiber.Ctx) error {
		if err := Status(c, 201); err != nil {
			return err
		}
		return JSON(c, 202, map[string]int{"a": 1})
	})
	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/", nil))
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != 202 {
		t.Errorf("status = %d, want 202", resp.StatusCode)
	}
}

func readBody(resp *http.Response) string {
	data := []byte{}
	buf := make([]byte, 4096)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			data = append(data, buf[:n]...)
		}
		if err != nil {
			break
		}
	}
	defer resp.Body.Close()
	return strings.TrimSpace(string(data))
}