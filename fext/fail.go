package fext

import (
	"fmt"
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

// ErrorHandler is optimized error handler impl
func ErrorHandler(c fiber.Ctx, err error) error {
	// Default status and structure
	status := fiber.StatusBadRequest
	resp := Fail{}

	// Optional conversion: transform raw error into *Fail
	err = convertError(err)

	// Type switch to handle different error categories
	switch e := err.(type) {
	case *Fail:
		if e.Status != 0 {
			status = e.Status
		}
		resp.Code = e.Code
		resp.Message = e.Message

		// A *Fail with neither a status nor a message carries no developer
		// intent, so treat it as an internal failure rather than echoing an
		// empty body with the default 400.
		if e.Status == 0 && e.Message == "" {
			status = fiber.StatusInternalServerError
			resp.Message = "internal server error"
		}

		// Handle system error visibility and logging
		if e.System != nil {
			if narder.IsDebug() {
				// Convert error to string for reliable JSON serialization
				if sysErr, ok := e.System.(error); ok {
					resp.System = sysErr.Error()
				} else {
					resp.System = e.System
				}
			}
			if sysErr, ok := e.System.(error); ok {
				narder.Error(sysErr)
			}
		}

	case *fiber.Error:
		if e.Code != 0 {
			status = e.Code
		}
		// Messages attached to 5xx errors are usually internal detail (driver
		// errors, upstream failures); only client errors are safe to echo.
		if e.Code >= 500 {
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
		narder.Error(fmt.Sprintf("%s %s: %v", c.Method(), c.Path(), err))
	}
	return c.Status(status).JSON(resp)
}
