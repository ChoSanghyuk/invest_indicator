package handler

/***************************************************************** request ****************************************************************/

type AssetHistReq struct {
	ID uint `json:"id" validate:"required"`
}

type TotalStatusResp struct {
	ID     uint    `json:"id"`
	Name   string  `json:"name"`
	Amount float64 `json:"amount"`
}

type AddFundReq struct {
	Name string `json:"name" validate:"required"`
}

type AddAssetReq struct {
	Name      string  `json:"name" validate:"required"`
	Category  uint    `json:"category" validate:"required,category"`
	Code      string  `json:"code"`
	Currency  string  `json:"currency" validate:"required"`
	Top       float64 `json:"top"`
	Bottom    float64 `json:"bottom"`
	Ema       float64 `json:"ema"`
	SellPrice float64 `json:"sel_price"`
	BuyPrice  float64 `json:"buy_price"`
}

type UpdateAssetReq struct {
	ID        uint    `json:"id" validate:"required"`
	Name      string  `json:"name"`
	Category  uint    `json:"category"`
	Code      string  `json:"code"`
	Currency  string  `json:"currency"`
	Top       float64 `json:"top"`
	Bottom    float64 `json:"bottom"`
	SellPrice float64 `json:"sel_price"`
	BuyPrice  float64 `json:"buy_price"`
}

type DeleteAssetReq struct {
	ID uint `json:"id" validate:"required"`
}

type SaveMarketStatusParam struct {
	Status uint `json:"status" validate:"required,market_status"`
}

type SaveInvestParam struct {
	FundId    uint    `json:"fund_id" validate:"required"`
	AssetId   uint    `json:"asset_id"`
	AssetName string  `json:"name"`
	AssetCode string  `json:"code"`
	Price     float64 `json:"price" validate:"required"`
	Count     float64 `json:"count" validate:"required"`
}

/***************************************************************** resoponse ****************************************************************/

type assetListResponse struct {
	AssetId   uint   `json:"asset_id"`
	AssetName string `json:"name"`
}

type assetResponse struct {
	ID        uint
	Name      string
	Category  string
	Code      string
	Currency  string
	Top       float64
	Bottom    float64
	SellPrice float64
	BuyPrice  float64
}

type HistResponse struct {
	FundId    uint    `json:"fund_id"`
	AssetId   uint    `json:"asset_id"`
	AssetName string  `json:"asset_name"`
	Count     float64 `json:"count"`
	Price     float64 `json:"price"`
	CreatedAt string  `json:"created_at"`
}

type fundAssetsResponse struct {
	FundId    uint    `json:"fund_id"`
	AssetId   uint    `json:"asset_id"`
	AssetName string  `json:"asset_name"`
	Count     float64 `json:"count"`
	Sum       float64 `json:"sum"`
}
