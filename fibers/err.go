package fibers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/natholdallas/natools4go/arrs"
)

var (
	errmode    bool                   = false
	errptr     func(err error)        = nil
	errhandler func(err error) *Error = nil
)

func SetErrorHandlerDevMode(s bool) {
	errmode = s
}

func SetErrorHandlerPrinter(s func(err error)) {
	errptr = s
}

func SetErrorHandler(s func(err error) *Error) {
	errhandler = s
}

type Error struct {
	Status  int    `json:"-"`                 // http status code
	Code    string `json:"code,omitempty"`    // business status code (optional)
	Message string `json:"message,omitempty"` // message (optional)
	System  error  `json:"system,omitempty"`  // system error (optional)
}

func (e Error) Error() string {
	return e.Message
}

func Err(status int, code, message string, system ...error) *Error {
	sys := arrs.GetDefault(nil, system...)
	return &Error{status, code, message, sys}
}

func BadRequest(code, message string, system ...error) *Error {
	sys := arrs.GetDefault(nil, system...)
	return &Error{fiber.StatusBadRequest, code, message, sys}
}

func Unauthorized(code, message string, system ...error) *Error {
	sys := arrs.GetDefault(nil, system...)
	return &Error{fiber.StatusUnauthorized, code, message, sys}
}

func Forbidden(code, message string, system ...error) *Error {
	sys := arrs.GetDefault(nil, system...)
	return &Error{fiber.StatusForbidden, code, message, sys}
}

func NotFound(code, message string, system ...error) *Error {
	sys := arrs.GetDefault(nil, system...)
	return &Error{fiber.StatusNotFound, code, message, sys}
}

// ErrorHandler is optimized error handler impl
func ErrorHandler(c *fiber.Ctx, err error) error {
	status := fiber.StatusBadRequest
	data := Error{}

	if errhandler != nil {
		err = errhandler(err)
	}

	switch e := err.(type) {
	case *Error:
		if e.Status != 0 {
			status = e.Status
		}
		data.Code = e.Code
		data.Message = e.Message
		if errmode {
			data.System = e.System
		}
		if errptr != nil {
			errptr(e.System)
		}

	case *fiber.Error:
		if e.Code != 0 {
			status = e.Code
		}
		data.Message = e.Message

	default:
		data.Message = e.Error()
	}

	return c.Status(status).JSON(data)
}
