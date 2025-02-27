package handler

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

type MarketHandler struct {
	r MaketRetriever
	w MarketSaver
}

func NewMarketHandler(r MaketRetriever, w MarketSaver) *MarketHandler {
	return &MarketHandler{
		r: r,
		w: w,
	}
}

func (h *MarketHandler) InitRoute(app *fiber.App) {

	router := app.Group("/market")
	router.Get("/:date?", h.Market)
	router.Get("/indicators/:date?", h.MarketIndicator)
	router.Get("/weekly_indicators", h.WeekMarketIndicators)
	router.Post("/", h.ChangeMarketStatus)
}

func (h *MarketHandler) Market(c *fiber.Ctx) error {

	date := c.Params("date")

	isDateFormat := dateCheck(date)
	if !isDateFormat {
		return fmt.Errorf("파라미터 유효성 검사 시 오류 발생. 올바르지 않은 date 포맷. %s", date)
	}

	market, err := h.r.RetrieveMarketStatus(date)
	if err != nil {
		return fmt.Errorf("RetrieveMarketStatus 오류 발생. %w", err)
	}

	return c.Status(fiber.StatusOK).JSON(market)
}

func (h *MarketHandler) MarketIndicator(c *fiber.Ctx) error {

	date := c.Params("date")

	isDateFormat := dateCheck(date)
	if !isDateFormat {
		return fmt.Errorf("파라미터 유효성 검사 시 오류 발생. 올바르지 않은 date 포맷. %s", date)
	}

	dailyIdx, cliIdx, err := h.r.RetrieveMarketIndicator(date)
	if err != nil {
		return fmt.Errorf("RetrieveMarketIndicator 오류 발생. %w", err)
	}

	return c.Status(fiber.StatusOK).JSON([]any{dailyIdx, cliIdx})
}

func (h *MarketHandler) WeekMarketIndicators(c *fiber.Ctx) error {

	weekIdx, err := h.r.RetrieveMarketIndicatorWeek()
	if err != nil {
		return fmt.Errorf("RetrieveMarketIndicatorWeek 오류 발생. %w", err)
	}

	fg := make([]uint, 7)
	nd := make([]float64, 7)

	for i, idx := range weekIdx {
		fg[i] = idx.FearGreedIndex
		nd[i] = idx.NasDaq
	}

	return c.Status(fiber.StatusOK).
		JSON(WeekMarketIndicators{
			FearGreedWeek: fg,
			NasdaqWeek:    nd,
		})
}

func (h *MarketHandler) ChangeMarketStatus(c *fiber.Ctx) error {

	var param SaveMarketStatusParam
	err := c.BodyParser(&param)
	if err != nil {
		return fmt.Errorf("파라미터 BodyParse 시 오류 발생. %w", err)
	}

	err = validCheck(&param)
	if err != nil {
		return fmt.Errorf("파라미터 유효성 검사 시 오류 발생. %w", err)
	}

	err = h.w.SaveMarketStatus(param.Status)
	if err != nil {
		return fmt.Errorf("RetrieveMarketStatus 오류 발생. %w", err)
	}

	return c.Status(fiber.StatusOK).SendString("시장 상태 저장 성공")

}
