package fext

import (
	"sync/atomic"

	"github.com/gofiber/fiber/v3"
)

var (
	debug   atomic.Bool
	errFunc atomic.Pointer[errorFunc]
)

type errorFunc struct{ fn func(err error) }

func SetDebugMode(mode bool) {
	debug.Store(mode)
}

func isDebug() bool {
	return debug.Load()
}

func SetErrorFunc(fn func(err error)) {
	if fn == nil {
		errFunc.Store(nil)
		return
	}
	errFunc.Store(&errorFunc{fn})
}

func onError(err error) {
	if e := errFunc.Load(); e != nil {
		e.fn(err)
	}
}

// Listen starts the server on the given address. If a listener has been
// registered via SetErrorFunc, the error is forwarded to it; otherwise it
// panics to make startup failures visible instead of silently dropping them.
func Listen(app *fiber.App, addr string, config ...fiber.ListenConfig) {
	if err := app.Listen(addr, config...); err != nil {
		if e := errFunc.Load(); e != nil {
			e.fn(err)
			return
		}
		panic(err)
	}
}
