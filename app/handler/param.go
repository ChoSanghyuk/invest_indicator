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
	Category  string  `json:"category" validate:"required,category"`
	Code      string  `json:"code"`
	Currency  string  `json:"currency" validate:"required"`
	Top       float64 `json:"top"`
	Bottom    float64 `json:"bottom"`
	SellPrice float64 `json:"sell"`
	BuyPrice  float64 `json:"buy"`
	Ema       float64 `json:"ema"`
	Ndays     uint    `json:"ndays"`
}

type UpdateAssetReq struct {
	ID        uint    `json:"id" validate:"required"`
	Name      string  `json:"name"`
	Category  string  `json:"category"`
	Code      string  `json:"code"`
	Currency  string  `json:"currency"`
	Top       float64 `json:"top"`
	Bottom    float64 `json:"bottom"`
	SellPrice float64 `json:"sell"`
	BuyPrice  float64 `json:"buy"`
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

type LoginRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

/***************************************************************** resoponse ****************************************************************/

type assetListResponse struct {
	AssetId   uint   `json:"asset_id"`
	AssetName string `json:"name"`
}

type assetResponse struct {
	ID        uint    `json:"id"`
	Name      string  `json:"name"`
	Category  string  `json:"category"`
	Code      string  `json:"code"`
	Currency  string  `json:"currency"`
	Top       float64 `json:"top"`
	Bottom    float64 `json:"bottom"`
	SellPrice float64 `json:"sell"`
	BuyPrice  float64 `json:"buy"`
	Ema       float64 `json:"ema"`
	NDays     float64 `json:"ndays"`
	Price     float64 `json:"price"`
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
	// FundId       uint    `json:"fund_id"`
	// AssetId      uint    `json:"asset_id"`
	// AssetName    string  `json:"asset_name"`
	// Count        float64 `json:"count"`
	// Sum          float64 `json:"sum"`
	Name         string `json:"name"`
	Amount       string `json:"amount"`
	AmountDollar string `json:"amount_dollar"`
	ProfitRate   string `json:"profit_rate"`
	Division     string `json:"division"`
	Quantity     string `json:"quantity"`
	Price        string `json:"price"`
	PriceDollar  string `json:"price_dollar"`
	IsStable     bool   `json:"isStable"`
}

type fundPortionResponse struct {
	Stable   int `json:"stable"`
	Volatile int `json:"volatile"`
}

type WeekMarketIndicators struct {
	FearGreedWeek []uint    `json:"fear_greed"`
	NasdaqWeek    []float64 `json:"nasdaq"`
	Sp500Week     []float64 `json:"sp500"`
}

// type FearGreedIndexResponse struct {
// 	Value      uint      `json:"value"`
// 	Status     string    `json:"status"`
// 	WeeklyData []float64 `json:"weeklyData"`
// }

// type NasdaqResponse struct {
// 	Value      float64   `json:"value"`
// 	Change     float64   `json:"change"`
// 	WeeklyData []float64 `json:"weeklyData"`
// }

// type SP500Response struct {
// 	Value      float64   `json:"value"`
// 	Change     float64   `json:"change"`
// 	WeeklyData []float64 `json:"weeklyData"`
// }

type EventResponse struct {
	Id          uint   `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Active      bool   `json:"active"`
}

type EventStatusChangeRequest struct {
	Id     uint `json:"id"`
	Active bool `json:"active"`
}

type EventLaunchRequest struct {
	Id uint `json:"id"`
}

// JWTResponse is the response sent after successful authentication
type JWTResponse struct {
	Token  string `json:"token"`
	Expiry int64  `json:"expiry"`
}
