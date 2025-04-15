package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

func SetupMiddleware(router fiber.Router) {

	router.Use(errorHandle)
}

func errorHandle(c *fiber.Ctx) error {

	err := c.Next()
	if err != nil {
		log.Error().Err(err).Msg("Error in middleware")
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}
	return nil
}
