package fext

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"

	"github.com/natholdallas/natools4go/narder"
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
	if resp.StatusCode != fiber.StatusInternalServerError {
		t.Errorf("status = %d, want 500", resp.StatusCode)
	}
	body := readBody(resp)
	if !strings.Contains(body, "internal server error") {
		t.Errorf("body = %s, want internal server error", body)
	}
	if strings.Contains(body, "boom") {
		t.Errorf("body leaked internal error text: %s", body)
	}
}

type plainErr struct{ msg string }

func (e *plainErr) Error() string { return e.msg }

func TestErrorHandlerSystemHiddenInProd(t *testing.T) {
	narder.SetDebugMode(false)
	defer narder.SetDebugMode(false)
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
	narder.SetDebugMode(true)
	defer narder.SetDebugMode(true)
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

func TestGetAuthorizationCaseInsensitiveNoLeadingSpace(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString(GetAuthorization(c))
	})
	for _, tc := range []struct{ in, want string }{
		{"bearer abc123", "abc123"},
		{"Bearer abc123", "abc123"},
		{"bearer  abc123", "abc123"}, // double space
		{"Bearer   abc123", "abc123"},
	} {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(fiber.HeaderAuthorization, tc.in)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test: %v", err)
		}
		if body := readBody(resp); body != tc.want {
			t.Errorf("GetAuthorization(%q) = %q, want %q", tc.in, body, tc.want)
		}
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

// TestGetAuthorizationRejectsForeignScheme guards against handing a different
// credential type to a token verifier: an "Authorization: Basic ..." header must
// not be returned as if it were the bearer token.
func TestGetAuthorizationRejectsForeignScheme(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString(GetAuthorization(c, "Token "))
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(fiber.HeaderAuthorization, "Bearer x")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if body := readBody(resp); body != "" {
		t.Errorf("authorization = %q, want empty for mismatched scheme", body)
	}
}

func TestErrorHandlerCodeOnlyIsClientError(t *testing.T) {
	narder.SetDebugMode(false)
	app := newTestApp()
	app.Post("/login", func(c fiber.Ctx) error {
		return &Fail{Code: "err.sign.in.failed"}
	})

	resp, err := app.Test(httptest.NewRequest(http.MethodPost, "/login", nil))
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
	body := readBody(resp)
	if !strings.Contains(body, `"code":"err.sign.in.failed"`) {
		t.Errorf("body = %s, want the error code", body)
	}
	if strings.Contains(body, "internal server error") {
		t.Errorf("code-only Fail must not be masked: %s", body)
	}
}

func TestErrorHandlerSystemWithoutStatusIsServerError(t *testing.T) {
	narder.SetDebugMode(false)
	app := newTestApp()
	app.Post("/create", func(c fiber.Ctx) error {
		return &Fail{Code: "err.data.create", System: &plainErr{"db down"}}
	})

	resp, err := app.Test(httptest.NewRequest(http.MethodPost, "/create", nil))
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != fiber.StatusInternalServerError {
		t.Errorf("status = %d, want 500", resp.StatusCode)
	}
	body := readBody(resp)
	if !strings.Contains(body, `"code":"err.data.create"`) || !strings.Contains(body, "internal server error") {
		t.Errorf("body = %s, want code plus masked message", body)
	}
	if strings.Contains(body, "db down") {
		t.Errorf("system detail leaked: %s", body)
	}
}

func TestErrorHandlerServerMessageMasked(t *testing.T) {
	narder.SetDebugMode(false)
	app := newTestApp()
	app.Get("/", func(c fiber.Ctx) error {
		return &Fail{Status: 500, Message: "sql: secret"}
	})

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/", nil))
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != fiber.StatusInternalServerError {
		t.Errorf("status = %d, want 500", resp.StatusCode)
	}
	if body := readBody(resp); strings.Contains(body, "sql: secret") {
		t.Errorf("5xx message leaked: %s", body)
	}
}

func TestErrorHandlerIllegalStatus(t *testing.T) {
	narder.SetDebugMode(false)
	app := newTestApp()
	app.Get("/", func(c fiber.Ctx) error {
		return &Fail{Status: fiber.StatusOK, Message: "x"}
	})

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/", nil))
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != fiber.StatusInternalServerError {
		t.Errorf("status = %d, want 500 for a non-error status", resp.StatusCode)
	}
}

func TestErrorHandlerEmptyFail(t *testing.T) {
	narder.SetDebugMode(false)
	app := newTestApp()
	app.Get("/", func(c fiber.Ctx) error {
		return &Fail{}
	})

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/", nil))
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != fiber.StatusInternalServerError {
		t.Errorf("status = %d, want 500 for an empty Fail", resp.StatusCode)
	}
}

func TestErrorHandlerLogsFiberServerError(t *testing.T) {
	narder.SetDebugMode(false)
	var buf bytes.Buffer
	narder.SetOutput(&buf)
	t.Cleanup(func() { narder.SetOutput(nil) })

	app := newTestApp()
	app.Get("/", func(c fiber.Ctx) error {
		return fiber.NewError(fiber.StatusInternalServerError, "db down")
	})

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/", nil))
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != fiber.StatusInternalServerError {
		t.Errorf("status = %d, want 500", resp.StatusCode)
	}
	if body := readBody(resp); strings.Contains(body, "db down") {
		t.Errorf("fiber 5xx message leaked: %s", body)
	}
	if logs := buf.String(); !strings.Contains(logs, "db down") {
		t.Errorf("logs = %q, want the masked message recorded", logs)
	}
}
