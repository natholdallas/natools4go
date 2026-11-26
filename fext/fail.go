package fext

import (
	"sync/atomic"

	"github.com/gofiber/fiber/v3"

	"github.com/natholdallas/natools4go/narder"
)

type Fail struct {
	Status  int    `json:"-"`                 // HTTP status code (not shown in body)
	Code    string `json:"code,omitempty"`    // Application-specific error code
	Message string `json:"message,omitempty"` // Human-readable error message
	System  any    `json:"system,omitempty"`  // Raw system error (only shown in debug mode)
} // @name Fail

func (e *Fail) Error() string {
	return e.Message
}

var errHandler atomic.Pointer[errorHandler]

type errorHandler struct{ fn func(err error) *Fail }

func SetErrorHandler(fn func(err error) *Fail) {
	if fn == nil {
		errHandler.Store(nil)
		return
	}
	errHandler.Store(&errorHandler{fn})
}

func convertError(err error) error {
	if h := errHandler.Load(); h != nil {
		if converted := h.fn(err); converted != nil {
			return converted
		}
	}
	return err
}

// ErrorHandler is optimized error handler impl.
//
// Status resolution:
//   - *Fail with an explicit Status uses it, except that a value outside
//     400..599 is a developer bug and becomes 500;
//   - *Fail without Status is 500 when it carries a System error (a server
//     fault) and 400 otherwise (a developer-authored client rejection);
//   - an entirely empty *Fail is 500: it carries no intent;
//   - *fiber.Error uses its Code;
//   - any other error is 500.
//
// A 5xx body never echoes internal text: Message is replaced with "internal
// server error" and the original is logged. Code survives, so clients can
// still localize the failure.
func ErrorHandler(c fiber.Ctx, err error) error {
	status := fiber.StatusBadRequest
	resp := Fail{}

	err = convertError(err)

	switch e := err.(type) {
	case *Fail:
		resp.Code = e.Code
		switch {
		case e.Status != 0:
			status = e.Status
		case e.System != nil:
			status = fiber.StatusInternalServerError
		case e.Code == "" && e.Message == "":
			status = fiber.StatusInternalServerError
		}
		if status < fiber.StatusBadRequest || status > 599 {
			status = fiber.StatusInternalServerError
		}
		if status >= fiber.StatusInternalServerError {
			if e.Message != "" {
				narder.Errorf("%s %s: %s", c.Method(), c.Path(), e.Message)
			}
			resp.Message = "internal server error"
		} else {
			resp.Message = e.Message
		}
		if e.System != nil {
			if narder.IsDebug() {
				// Convert error to string for reliable JSON serialization
				if sysErr, ok := e.System.(error); ok {
					resp.System = sysErr.Error()
				} else {
					resp.System = e.System
				}
			}
			narder.Errorf("%s %s: %v", c.Method(), c.Path(), e.System)
		}

	case *fiber.Error:
		if e.Code != 0 {
			status = e.Code
		}
		if status < fiber.StatusBadRequest || status > 599 {
			status = fiber.StatusInternalServerError
		}
		if status >= fiber.StatusInternalServerError {
			narder.Errorf("%s %s: %s", c.Method(), c.Path(), e.Message)
			resp.Message = "internal server error"
		} else {
			resp.Message = e.Message
		}

	default:
		// Unknown errors must never be serialized into the response body: they
		// routinely contain DSNs, SQL and stack detail. Log and return a fixed
		// message instead.
		status = fiber.StatusInternalServerError
		resp.Message = "internal server error"
		narder.Errorf("%s %s: %v", c.Method(), c.Path(), err)
	}
	return c.Status(status).JSON(resp)
}
