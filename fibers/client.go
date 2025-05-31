package fibers

import "github.com/gofiber/fiber/v2"

func Agent(s fiber.Agent) (int, []byte, error) {
	code, body, errs := s.Bytes()
	if len(errs) > 0 {
		return code, body, errs[0]
	}
	return code, body, nil
}
