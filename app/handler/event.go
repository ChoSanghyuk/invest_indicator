package handler

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

type EventHandler struct {
	er EventRetriever
	el EventLauncher
	ec EventStatusChanger
}

func NewEventHandler(er EventRetriever, el EventLauncher, ec EventStatusChanger) *EventHandler {
	return &EventHandler{
		er: er,
		el: el,
		ec: ec,
	}
}

func (h *EventHandler) InitRoute(app *fiber.App) {

	router := app.Group("/events")
	router.Get("/", h.Events)
	router.Post("/switch", h.SwtichEvent)
	router.Post("/launch", h.LaunchEvent)
}

func (h *EventHandler) Events(c *fiber.Ctx) error {

	events := h.er.Events()

	eventResponse := make([]EventResponse, 0, len(events))
	for _, e := range events {
		eventResponse = append(eventResponse, EventResponse{
			Id:          e.Id,
			Title:       e.Title,
			Description: e.Description,
			Active:      e.IsActive,
		})
	}

	return c.Status(fiber.StatusOK).JSON(eventResponse)
}

func (h *EventHandler) SwtichEvent(c *fiber.Ctx) error {

	param := EventStatusChangeRequest{}
	err := c.BodyParser(&param)
	if err != nil {
		return fmt.Errorf("파라미터 BodyParse 시 오류 발생. %w", err)
	}

	err = validCheck(&param)
	if err != nil {
		return fmt.Errorf("파라미터 유효성 검사 시 오류 발생. %w", err)
	}

	err = h.ec.SetEventStatus(param.Id, param.Active)
	if err != nil {
		return fmt.Errorf("상태 변경 요청 시 오류. %w", err)
	}

	return c.Status(fiber.StatusOK).SendString("Event 실행 성공")
}

func (h *EventHandler) LaunchEvent(c *fiber.Ctx) error {

	var param EventLaunchRequest
	err := c.BodyParser(&param)
	if err != nil {
		return fmt.Errorf("파라미터 BodyParse 시 오류 발생. %w", err)
	}

	err = validCheck(&param)
	if err != nil {
		return fmt.Errorf("파라미터 유효성 검사 시 오류 발생. %w", err)
	}

	err = h.el.LaunchEvent(param.Id)
	if err != nil {
		return fmt.Errorf("event Launch 시 오류 발생. %s", err.Error())
	}

	return c.Status(fiber.StatusOK).SendString("event 실행 성공")

}
