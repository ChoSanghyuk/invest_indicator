package handler

import (
	"invest/model"

	"github.com/gofiber/fiber/v2"
)

type CategoryHandler struct {
}

func (h *CategoryHandler) InitRoute(app *fiber.App) {
	router := app.Group("/categories")
	router.Get("", h.GetCategories)
}

func NewCategoryHandler() *CategoryHandler {
	return &CategoryHandler{}
}

func (h *CategoryHandler) GetCategories(c *fiber.Ctx) error {
	return c.JSON(model.CategoryList())
}
