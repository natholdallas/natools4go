package fibers

import (
	"github.com/gofiber/fiber/v2"
)

type Error struct {
	Status  int    `json:"-"`                 // http status code
	Code    string `json:"code,omitempty"`    // business status code (optional)
	Message string `json:"message,omitempty"` // message (optional)
	System  error  `json:"system,omitempty"`  // system error (optional)
}

func (e Error) Error() string {
	return e.Message
}

var (
	errDevMode bool                  = false
	errPrinter func(err error)       = nil
	errHandler func(err error) Error = nil
)

func SetErrorHandlerDevMode(s bool) {
	errDevMode = s
}

func SetErrorHandlerPrinter(s func(err error)) {
	errPrinter = s
}

func SetErrorHandler(s func(err error) Error) {
	errHandler = s
}

// ErrorHandler is optimized error handler impl
func ErrorHandler(c *fiber.Ctx, err error) error {
	status := fiber.StatusBadRequest
	data := Error{}

	switch e := err.(type) {
	case *Error:
		data.Code = e.Code
		data.Message = e.Message
		if e.Status != 0 {
			status = e.Status
		}
		if errDevMode {
			data.System = e.System
		}
		if errPrinter != nil {
			errPrinter(e.System)
		}

	case *fiber.Error:
		if e.Code != 0 {
			status = e.Code
		}
		data.Message = e.Message

	default:
		if errHandler != nil {
			v := errHandler(e)
			data.Code = v.Code
			data.Message = v.Message
			if v.Status != 0 {
				status = v.Status
			}
			if errDevMode {
				data.System = v.System
			}
			if errPrinter != nil {
				errPrinter(v.System)
			}
		}
		data.Message = e.Error()
	}
	return c.Status(status).JSON(data)
}
