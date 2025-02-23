package handler

import (
	"invest/model"

	"github.com/gofiber/fiber/v2"
)

type ModelHandler struct {
}

func (h *ModelHandler) InitRoute(app *fiber.App) {
	app.Get("/categories", h.GetCategories)
	app.Get("/currencies", h.GetCurrencies)
}

func NewCategoryHandler() *ModelHandler {
	return &ModelHandler{}
}

func (h *ModelHandler) GetCategories(c *fiber.Ctx) error {
	return c.JSON(model.CategoryList())
}

func (h *ModelHandler) GetCurrencies(c *fiber.Ctx) error {
	return c.JSON(model.CurrencyList())
}
