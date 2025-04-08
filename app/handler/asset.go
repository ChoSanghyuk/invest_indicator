package handler

import (
	"fmt"
	m "invest/model"
	"math"

	"github.com/gofiber/fiber/v2"
)

type AssetHandler struct {
	r AssetRetriever
	w AssetInfoSaver
	p PriceGetter
}

func NewAssetHandler(r AssetRetriever, w AssetInfoSaver, p PriceGetter) *AssetHandler {
	return &AssetHandler{
		r: r,
		w: w,
		p: p,
	}
}

func (h *AssetHandler) InitRoute(app *fiber.App) {

	router := app.Group("/assets")

	router.Post("/", h.AddAsset)
	router.Put("/", h.UpdateAsset)
	router.Delete("/", h.DeleteAsset)
	router.Get("", h.Assets)
	router.Get("/list", h.AssetList)
	router.Get("/:id<\\d+>", h.Asset)
	router.Get("/:id<\\d+>/hist", h.AssetHist)
}

func (h *AssetHandler) Assets(c *fiber.Ctx) error {

	assets, err := h.r.RetrieveAssetList()
	if err != nil {
		return fmt.Errorf("RetrieveAssetList 오류 발생. %w", err)
	}

	rtn := make([]assetResponse, len(assets))

	for i, a := range assets {
		asset, err := h.r.RetrieveAsset(a.ID)
		if err != nil {
			return fmt.Errorf("RetrieveAsset 오류 발생. %w", err)
		}

		pp, err := h.p.PresentPrice(asset.Category, asset.Code)
		if err != nil {
			// todo. log
			pp = 0
		}

		rtn[i] = assetResponse{
			ID:        asset.ID,
			Name:      asset.Name,
			Category:  asset.Category.String(),
			Code:      asset.Code,
			Currency:  asset.Currency,
			Top:       asset.Top,
			Bottom:    asset.Bottom,
			SellPrice: asset.SellPrice,
			BuyPrice:  asset.BuyPrice,
			Price:     pp,
		}

	}

	return c.Status(fiber.StatusOK).JSON(rtn)
}

func (h *AssetHandler) AddAsset(c *fiber.Ctx) error {

	var param AddAssetReq
	err := c.BodyParser(&param)
	if err != nil {
		return fmt.Errorf("파라미터 BodyParse 시 오류 발생. %w", err)
	}

	err = validCheck(&param)
	if err != nil {
		return fmt.Errorf("파라미터 유효성 검사 시 오류 발생. %w", err)
	}
	category, err := m.ToCategory(param.Category)
	if err != nil {
		return fmt.Errorf("카테고리 변환 시 오류 발생. %w", err)
	}

	top, bottom := param.Top, param.Bottom
	if top == 0 || bottom == 0 {
		_top, _bottom, err := h.p.TopBottomPrice(category, param.Code)
		if err != nil {
			// return fmt.Errorf("TopBottomPrice 시 오류 발생. %w", err) todo.log
			top = 0
			bottom = math.MaxInt32
		}
		if top == 0 {
			top = _top
		}
		if bottom == 0 {
			bottom = _bottom
		}
	}

	if param.BuyPrice == 0 {
		param.BuyPrice = bottom
	}

	id, err := h.w.SaveAssetInfo(m.Asset{
		Name:      param.Name,
		Category:  category,
		Code:      param.Code,
		Currency:  param.Currency,
		Top:       top,
		Bottom:    bottom,
		SellPrice: param.SellPrice,
		BuyPrice:  param.BuyPrice,
	})
	if err != nil {
		return fmt.Errorf("SaveAssetInfo 시 오류 발생. %w", err)
	}

	var ema float64
	var n uint

	if param.Ema == 0 {
		ema, n, _ = h.p.AvgPrice(category, param.Code)
	} else {
		ema = param.Ema
		n = param.Ndays
	}

	err = h.w.SaveEmaHist(&m.EmaHist{
		AssetID: id,
		Ema:     ema,
		NDays:   n,
	})
	if err != nil {
		return fmt.Errorf("SaveEmaHist 시 오류 발생. %w", err)
	}

	return c.Status(fiber.StatusOK).SendString("자산 정보 저장 성공")
}

func (h *AssetHandler) UpdateAsset(c *fiber.Ctx) error {

	var param UpdateAssetReq
	err := c.BodyParser(&param)
	if err != nil {
		return fmt.Errorf("파라미터 BodyParse 시 오류 발생. %w", err)
	}

	err = validCheck(&param) // 포인터로 들어가도 validation 체크 되는지 확인
	if err != nil {
		return fmt.Errorf("파라미터 유효성 검사 시 오류 발생. %w", err)
	}

	category, err := m.ToCategory(param.Category)
	if err != nil {
		return fmt.Errorf("카테고리 변환 시 오류 발생. %w", err)
	}

	err = h.w.UpdateAssetInfo(m.Asset{
		ID:        param.ID,
		Name:      param.Name,
		Category:  category,
		Code:      param.Code,
		Currency:  param.Currency,
		Top:       param.Top,
		Bottom:    param.Bottom,
		SellPrice: param.SellPrice,
		BuyPrice:  param.BuyPrice,
	})
	if err != nil {
		return fmt.Errorf("UpdateAssetInfo 시 오류 발생. %w", err)
	}

	return c.Status(fiber.StatusOK).SendString("자산 정보 갱신 성공")
}

func (h *AssetHandler) DeleteAsset(c *fiber.Ctx) error {

	var param DeleteAssetReq
	err := c.BodyParser(&param)
	if err != nil {
		return fmt.Errorf("파라미터 BodyParse 시 오류 발생. %w", err)
	}

	err = validCheck(&param) // 포인터로 들어가도 validation 체크 되는지 확인
	if err != nil {
		return fmt.Errorf("파라미터 유효성 검사 시 오류 발생. %w", err)
	}

	err = h.w.DeleteAssetInfo(param.ID)
	if err != nil {
		return fmt.Errorf("DeleteAssetInfo 시 오류 발생. %w", err)
	}

	return c.Status(fiber.StatusOK).SendString("자산 정보 삭제 성공")
}

func (h *AssetHandler) Asset(c *fiber.Ctx) error {

	id, err := c.ParamsInt("id")
	if err != nil {
		return fmt.Errorf("파라미터 id 조회 시 오류 발생. %w", err)
	}

	asset, err := h.r.RetrieveAsset(uint(id))
	if err != nil {
		return fmt.Errorf("RetrieveAsset 오류 발생. %w", err)
	}

	rtn := assetResponse{
		ID:        asset.ID,
		Name:      asset.Name,
		Category:  asset.Category.String(),
		Code:      asset.Code,
		Currency:  asset.Currency,
		Top:       asset.Top,
		Bottom:    asset.Bottom,
		SellPrice: asset.SellPrice,
		BuyPrice:  asset.BuyPrice,
	}

	return c.Status(fiber.StatusOK).JSON(rtn)
}

func (h *AssetHandler) AssetList(c *fiber.Ctx) error {
	assets, err := h.r.RetrieveAssetList()
	if err != nil {
		return fmt.Errorf("RetrieveAssetList 오류 발생. %w", err)
	}
	rtn := make([]assetListResponse, len(assets)) // memo. pointer slice가 아니라도 값 변경
	for i, a := range assets {
		rtn[i].AssetId = a.ID
		rtn[i].AssetName = a.Name
	}

	return c.Status(fiber.StatusOK).JSON(rtn)
}

func (h *AssetHandler) AssetHist(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return fmt.Errorf("파라미터 id 조회 시 오류 발생. %w", err)
	}

	hist, err := h.r.RetrieveAssetHist(uint(id))
	if err != nil {
		return fmt.Errorf("RetrieveAssetHist 오류 발생. %w", err)
	}

	resp := make([]HistResponse, len(hist))
	for i, h := range hist {
		resp[i] = HistResponse{
			FundId:    h.FundID,
			AssetId:   h.AssetID,
			AssetName: h.Asset.Name,
			Count:     h.Count,
			Price:     h.Price,
			CreatedAt: h.CreatedAt.Format("20060102"),
		}
	}

	return c.Status(fiber.StatusOK).JSON(resp)

}
