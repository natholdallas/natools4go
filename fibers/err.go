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
	errmode    bool                  = false
	errptr     func(err error)       = nil
	errhandler func(err error) Error = nil
)

func SetErrorHandlerDevMode(s bool) {
	errmode = s
}

func SetErrorHandlerPrinter(s func(err error)) {
	errptr = s
}

func SetErrorHandler(s func(err error) Error) {
	errhandler = s
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
		if errhandler != nil {
			v := errhandler(e)
			data.Code = v.Code
			data.Message = v.Message
			if v.Status != 0 {
				status = v.Status
			}
			if errmode {
				data.System = v.System
			}
			if errptr != nil {
				errptr(v.System)
			}
		} else {
			data.Message = e.Error()
		}
	}
	return c.Status(status).JSON(data)
}
