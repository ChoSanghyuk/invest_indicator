package middleware

import "github.com/gofiber/fiber/v2"

func SetupMiddleware(router fiber.Router) {

	router.Use(errorHandle)
}

func errorHandle(c *fiber.Ctx) error {

	err := c.Next()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}
	return nil
}
