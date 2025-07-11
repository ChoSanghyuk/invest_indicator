package handler

import (
	"fmt"
	"invest/model"

	"github.com/gofiber/fiber/v2"
)

type FundHandler struct {
	r FundRetriever
	w FundWriter
	i InvestRetriever
	e ExchageRateGetter
}

func NewFundHandler(r FundRetriever, w FundWriter, i InvestRetriever, e ExchageRateGetter) *FundHandler {
	return &FundHandler{
		r: r,
		w: w,
		i: i,
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

	invests, err := h.r.RetreiveFundSummaryByFundId(uint(id))
	if err != nil {
		return fmt.Errorf("RetreiveFundSummaryById 시 오류 발생. %w", err)
	}

	resp := make([]fundAssetsResponse, 0, len(invests))

	for _, iv := range invests {
		if iv.Count == 0 {
			continue
		}

		fundAsset := fundAssetsResponse{
			Name: iv.Asset.Name,
			// ProfitRate: "", // todo ProfitRate 계산 로직 추가
			Division: iv.Asset.Category.String(),
			Quantity: fmt.Sprintf("%.2f", iv.Count),
			IsStable: iv.Asset.Category.IsStable(),
		}

		if iv.Asset.Currency == model.KRW.String() {
			fundAsset.Amount = fmt.Sprintf("%.2f", iv.Sum)
		} else {
			fundAsset.Amount = fmt.Sprintf("%.2f", iv.Sum*h.e.ExchageRate())
			fundAsset.AmountDollar = fmt.Sprintf("%.2f", iv.Sum)
		}

		fundAsset.ProfitRate = h.profitRateOfAsset(&iv)
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

	start := c.Query("start")
	end := c.Query("end")

	var invests []model.Invest

	if start == "" || end == "" {
		invests, err = h.r.RetreiveFundInvestsById(uint(id))
	} else {
		invests, err = h.r.RetrieveFundInvestsByIdAndRange(uint(id), start, end)
	}
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
			CreatedAt: iv.CreatedAt.Format("2006-01-02 15:04:05"),
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
		v := 0.0
		if f.Asset.Currency == model.KRW.String() {
			v = f.Sum
		} else {
			v = f.Sum * h.e.ExchageRate()

		}

		if f.Asset.Category.IsStable() {
			stableAmount += v
		} else {
			volatileAmount += v
		}
	}

	resp := fundPortionResponse{}
	if stableAmount != 0 || volatileAmount != 0 {
		volatilePortion := int((volatileAmount / (stableAmount + volatileAmount)) * 100)
		resp.Volatile = volatilePortion
		resp.Stable = 100 - volatilePortion
	}

	return c.Status(fiber.StatusOK).JSON(resp)

}

// 수익률 = (현재가치 + 판매가치) / 총 구입 가치
func (h *FundHandler) profitRateOfAsset(iv *model.InvestSummary) string {
	if iv.Asset.Category == model.Won || iv.Asset.Category == model.Dollar {
		return ""
	}

	invests, err := h.i.RetrieveInvestHist(iv.FundID, iv.AssetID, "", "")
	if err != nil {
		return ""
	}

	// base := 0.0
	// sold := 0.0
	// for i := len(invests) - 1; i >= 0; i-- {
	// 	invest := invests[i]
	// 	if invest.Count < 0 {
	// 		sold += invest.Count * -1
	// 		continue
	// 	}
	// 	if sold > 0 {
	// 		if sold > invest.Count {
	// 			sold -= invest.Count
	// 		} else {
	// 			invest.Count -= sold
	// 			sold = 0
	// 		}
	// 	}
	// 	base += invest.Count * invest.Price
	// }
	// if base == 0 {
	// 	return ""
	// }
	// return fmt.Sprintf("%.2f", 100*(iv.Sum-base)/base)

	soldValue := 0.0
	buyValue := 0.0

	for _, iv := range invests {
		if iv.Count < 0 {
			soldValue += iv.Count * -1 * iv.Price
		} else {
			buyValue += iv.Count * iv.Price
		}
	}

	if buyValue == 0 {
		return ""
	}

	return fmt.Sprintf("%.2f", 100*(iv.Sum-buyValue)/buyValue)
}
