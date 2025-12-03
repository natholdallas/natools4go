package fibers

import (
	"github.com/gofiber/fiber/v2"
)

/*
func FindUser(c *fiber.Ctx) error {
	 user, err := db.FindUserByID(1)
	 if err != nil {
	 	 return Error{
	 	 	 Status:  fiber.StatusBadRequest,
	 	 	 Code:    "err.record.not.found",
	 	 	 Message: "record not found in our database",
	 	 	 System:  err,
	 	 }
	 }
	 return nil
}
*/

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
		status = e.Status
		data.Code = e.Code
		data.Message = e.Message
		if errDevMode {
			data.System = e.System
		}
		if errPrinter != nil {
			errPrinter(e.System)
		}

	case *fiber.Error:
		status = e.Code
		data.Message = e.Message

	default:
		if errHandler != nil {
			data = errHandler(e)
			status = data.Status
		} else {
			data.Message = e.Error()
		}
	}
	return c.Status(status).JSON(data)
}
