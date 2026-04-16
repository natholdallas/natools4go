package fext

import (
	"github.com/gofiber/fiber/v3"
)

var ErrorConv func(err error) *Fail = nil

type Fail struct {
	Status  int    `json:"-"`                 // HTTP status code (not shown in body)
	Code    string `json:"code,omitempty"`    // Application-specific error code
	Message string `json:"message,omitempty"` // Human-readable error message
	System  any    `json:"system,omitempty"`  // Raw system error (only shown in debug mode)
} // @name Fail

func (e *Fail) Error() string {
	return e.Message
}

// ErrorHandler is optimized error handler impl
func ErrorHandler(c fiber.Ctx, err error) error {
	// Default status and structure
	status := fiber.StatusBadRequest
	resp := Fail{}

	// Optional conversion: transform raw error into *Error
	if ErrorConv != nil {
		if converted := ErrorConv(err); converted != nil {
			err = converted
		}
	}

	// Type switch to handle different error categories
	switch e := err.(type) {
	case *Fail:
		if e.Status != 0 {
			status = e.Status
		}
		resp.Code = e.Code
		resp.Message = e.Message

		// Handle system error visibility and logging
		if e.System != nil {
			if DebugMode {
				// Convert error to string for reliable JSON serialization
				if sysErr, ok := e.System.(error); ok {
					resp.System = sysErr.Error()
				} else {
					resp.System = e.System
				}
			}
			if ErrorFunc != nil {
				if sysErr, ok := e.System.(error); ok {
					ErrorFunc(sysErr)
				}
			}
		}

	case *fiber.Error:
		if e.Code != 0 {
			status = e.Code
		}
		resp.Message = e.Message

	default:
		resp.Message = err.Error()
	}
	return c.Status(status).JSON(resp)
}
