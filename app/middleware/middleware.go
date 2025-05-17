package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/rs/zerolog/log"
)

func SetupMiddleware(router fiber.Router) {

	router.Use(cors.New(cors.Config{
		AllowOrigins: "http://www.lomoninvest.shop:50001", // Replace with your Flutter web origin
		AllowHeaders: "Origin, Content-Type, Authorization",
		AllowMethods: strings.Join([]string{
			fiber.MethodGet,
			fiber.MethodPost,
			fiber.MethodHead,
			fiber.MethodPut,
			fiber.MethodDelete,
			fiber.MethodPatch,
		}, ","),
		AllowCredentials: true,
		MaxAge:           60 * 60 * 1,
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
