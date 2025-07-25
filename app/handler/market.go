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
	router.Get("/weekly_indicators", h.WeekMarketIndicators)
	router.Get("/indicators/:date?", h.MarketIndicator)
	router.Get("/:date?", h.Market)
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

	weekIdx, err := h.r.RetrieveMarketIndicatorWeekDesc()
	if err != nil {
		return fmt.Errorf("RetrieveMarketIndicatorWeek 오류 발생. %w", err)
	}

	l := len(weekIdx)
	fgw := make([]float64, l)
	ndw := make([]float64, l)
	spw := make([]float64, l)

	for i, idx := range weekIdx {
		fgw[l-(i+1)] = float64(idx.FearGreedIndex)
		ndw[l-(i+1)] = idx.NasDaq
		spw[l-(i+1)] = idx.Sp500
	}

	// fg
	fg := MarketIndexInner{
		Value:     fmt.Sprintf("%.0f", fgw[l-1]),
		GraphData: fgw,
	}
	if fgw[l-1] > 50 {
		fg.Status = "GREED"
	} else {
		fg.Status = "FEAR"
	}

	// nd
	nd := MarketIndexInner{
		Value:     fmt.Sprintf("%.2f", ndw[l-1]),
		Status:    fmt.Sprintf("%.2f", 100*(ndw[l-1]-ndw[l-2])/ndw[l-1]) + "%",
		GraphData: ndw,
	}

	// sp
	sp := MarketIndexInner{
		Value:     fmt.Sprintf("%.2f", spw[l-1]),
		Status:    fmt.Sprintf("%.2f", 100*(spw[l-1]-spw[l-2])/spw[l-1]) + "%",
		GraphData: spw,
	}

	hyw, err := h.r.RetrieveHighYieldSpreadWeekDesc()
	if err != nil {
		return fmt.Errorf("RetrieveHighYieldSpread 오류 발생. %w", err)
	}

	hy := MarketIndexInner{
		Value:     fmt.Sprintf("%.2f", hyw[l-1]),
		Status:    fmt.Sprintf("%.2f", 100*(hyw[l-1]-hyw[l-2])/hyw[l-1]) + "%",
		GraphData: hyw,
	}

	return c.Status(fiber.StatusOK).
		JSON(map[string]MarketIndexInner{
			"Fear & Greed Index": fg,
			"NASDAQ":             nd,
			"S&P 500":            sp,
			"High Yield Spread":  hy,
		})
}

/*
	{
	'Fear & Greed Index': {
		'value': fearGreedWeek[6].toString(),
		'status': fearGreedWeek[6] > 50? 'GREED' : 'FEAR',
		'graph': fearGreedWeek
	},
	'NASDAQ' : {
	'value': nasdaqWeek[6].toString(),
	'status': (100 * ((nasdaqWeek[6]-nasdaqWeek[5]) / nasdaqWeek[6])).toStringAsFixed(2)+'%' ,
	'graph': nasdaqWeek
	},
	'S&P 500' : {
	'value': sp500Week[6].toString(),
	'status': (100 * ((sp500Week[6]-sp500Week[5]) / sp500Week[6])).toStringAsFixed(2)+'%' ,
	'graph': sp500Week
	}
};
*/

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
