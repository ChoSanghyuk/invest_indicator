package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

func SetupMiddleware(router fiber.Router) {

	router.Use(cors.New(cors.Config{
    AllowOrigins: "http://localhost:50002", // temp. Flutter web server address
    AllowMethods: "GET,POST,PUT,DELETE",
	}))
	router.Use(errorHandle)
	router.Use(logRequest)
}

func errorHandle(c *fiber.Ctx) error {

	err := c.Next()
	if err != nil {
		log.Error().Err(err).Msg("Error in middleware")
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}
	return nil
}

func logRequest(c *fiber.Ctx) error {
	log.Info().Str("endpoint", c.Path()).Msg("Request endpoint")
	log.Info().Str("body", string(c.Body())).Msg("Request body")
	return c.Next()
}
