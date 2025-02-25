package handler

import (
	"fmt"
	"invest/model"

	"github.com/gofiber/fiber/v2"
)

type FundHandler struct {
	r FundRetriever
	w FundWriter
	e ExchageRateGetter
}

func NewFundHandler(r FundRetriever, w FundWriter, e ExchageRateGetter) *FundHandler {
	return &FundHandler{
		r: r,
		w: w,
		e: e,
	}
}

func (h *FundHandler) InitRoute(app *fiber.App) {
	router := app.Group("/funds")

	router.Get("/", h.TotalStatus)
	router.Post("/", h.AddFund)
	router.Get("/:id/hist", h.FundHist)
	router.Get("/:id/assets", h.FundAssets)
	router.Get("/:id/portion", h.FundPortion)
}

// 총 자금 금액
func (h *FundHandler) TotalStatus(c *fiber.Ctx) error {

	var exchangeRate float64 = h.e.ExchageRate()

	investSummarys, err := h.r.RetreiveFundsSummaryOrderByFundId()
	if err != nil {
		return fmt.Errorf("RetreiveFundSummary 오류 발생. %w", err)
	}

	funds := make(map[uint]*TotalStatusResp)
	for _, is := range investSummarys {

		if funds[is.FundID] == nil {
			funds[is.FundID] = &TotalStatusResp{
				ID:   is.FundID,
				Name: is.Fund.Name,
			}
		}

		if is.Asset.Currency == model.USD.String() {
			funds[is.FundID].Amount += is.Sum * exchangeRate
		} else {
			funds[is.FundID].Amount += is.Sum
		}
	}

	return c.Status(fiber.StatusOK).JSON(funds)
}

func (h *FundHandler) AddFund(c *fiber.Ctx) error {

	var param AddFundReq
	err := c.BodyParser(&param)
	if err != nil {
		return fmt.Errorf("파라미터 BodyParse 시 오류 발생. %w", err)
	}

	err = validCheck(param) // 포인터로 들어가도 validation 체크 되는지 확인
	if err != nil {
		return fmt.Errorf("파라미터 유효성 검사 시 오류 발생. %w", err)
	}

	err = h.w.SaveFund(param.Name)
	if err != nil {
		return fmt.Errorf("SaveFund 시 오류 발생. %w", err)
	}

	return c.Status(fiber.StatusOK).SendString("자금 정보 저장 성공")
}

// 자금별 보유 자산
func (h *FundHandler) FundAssets(c *fiber.Ctx) error {

	id, err := c.ParamsInt("id")
	if err != nil {
		return fmt.Errorf("파라미터 id 조회 시 오류 발생. %w", err)
	}

	funds, err := h.r.RetreiveFundSummaryByFundId(uint(id))
	if err != nil {
		return fmt.Errorf("RetreiveFundSummaryById 시 오류 발생. %w", err)
	}

	resp := make([]fundAssetsResponse, 0, len(funds))

	for _, f := range funds {
		if f.Count == 0 {
			continue
		}

		fundAsset := fundAssetsResponse{
			Name:        f.Asset.Name,
			ProfitRate:  "", // todo ProfitRate 계산 로직 추가
			Division:    f.Asset.Category.String(),
			Quantity:    fmt.Sprintf("%f", f.Count),
			Price:       "",
			PriceDollar: "",
			IsStable:    f.Asset.Category.IsStable(),
		}
		if f.Asset.Currency == model.Won.String() {
			fundAsset.Amount = fmt.Sprintf("%f", f.Sum)
		} else {
			fundAsset.Amount = fmt.Sprintf("%f", f.Sum*h.e.ExchageRate())
			fundAsset.AmountDollar = fmt.Sprintf("%f", f.Sum)

		}

		resp = append(resp, fundAsset)
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}

// 자금별 투자 이력
func (h *FundHandler) FundHist(c *fiber.Ctx) error {

	id, err := c.ParamsInt("id")
	if err != nil {
		return fmt.Errorf("파라미터 id 조회 시 오류 발생. %w", err)
	}

	invests, err := h.r.RetreiveAFundInvestsById(uint(id))
	if err != nil {
		return fmt.Errorf("RetreiveAFundInvestsById 시 오류 발생. %w", err)
	}
	fundHists := make([]HistResponse, len(invests))
	for i, iv := range invests {
		fundHists[i] = HistResponse{
			FundId:    iv.FundID,
			AssetId:   iv.AssetID,
			AssetName: iv.Asset.Name,
			Count:     iv.Count,
			Price:     iv.Price,
			CreatedAt: iv.CreatedAt.Format("20060102"),
		}
	}

	return c.Status(fiber.StatusOK).JSON(fundHists)
}

func (h *FundHandler) FundPortion(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return fmt.Errorf("파라미터 id 조회 시 오류 발생. %w", err)
	}

	funds, err := h.r.RetreiveFundSummaryByFundId(uint(id))
	if err != nil {
		return fmt.Errorf("RetreiveFundSummaryById 시 오류 발생. %w", err)
	}

	stableAmount := 0.0
	volatileAmount := 0.0

	for _, f := range funds {
		if f.Count == 0 {
			continue
		}

		if f.Asset.Category.IsStable() {
			stableAmount += f.Sum
		} else {
			volatileAmount += f.Sum
		}
	}

	resp := fundPortionResponse{}
	if stableAmount != 0 || volatileAmount != 0 {
		stablePortion := int((stableAmount / (stableAmount + volatileAmount)) * 100)
		resp.Stable = stablePortion
		resp.Volatile = 100 - stablePortion
	}

	return c.Status(fiber.StatusOK).JSON(resp)

}
