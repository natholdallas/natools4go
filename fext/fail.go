package fext

import (
	"github.com/gofiber/fiber/v2"
)

var (
	debugMode    bool                  = false
	errLogger    func(err error)       = nil
	errConverter func(err error) *Fail = nil
)

// SetDebugMode enables or disables the exposure of system-level errors in the response.
func SetDebugMode(enabled bool) {
	debugMode = enabled
}

// SetErrorLogger registers a custom function to log internal system errors.
// error never be nil
func SetErrorLogger(fn func(err error)) {
	errLogger = fn
}

// SetErrorConverter registers a function to transform generic errors into the custom Error type.
func SetErrorConverter(fn func(err error) *Fail) {
	errConverter = fn
}

type Fail struct {
	Status  int    `json:"-"`                 // HTTP status code (not shown in body)
	Code    string `json:"code,omitempty"`    // Application-specific error code
	Message string `json:"message,omitempty"` // Human-readable error message
	System  any    `json:"system,omitempty"`  // Raw system error (only shown in debug mode)
}

func (e *Fail) Error() string {
	return e.Message
}

// ErrorHandler is optimized error handler impl
func ErrorHandler(c *fiber.Ctx, err error) error {
	// Default status and structure
	status := fiber.StatusBadRequest
	resp := Fail{}

	// Optional conversion: transform raw error into *Error
	if errConverter != nil {
		if converted := errConverter(err); converted != nil {
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
			if debugMode {
				// Convert error to string for reliable JSON serialization
				if sysErr, ok := e.System.(error); ok {
					resp.System = sysErr.Error()
				} else {
					resp.System = e.System
				}
			}
			if errLogger != nil {
				if sysErr, ok := e.System.(error); ok {
					errLogger(sysErr)
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
