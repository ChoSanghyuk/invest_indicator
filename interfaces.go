package investind

import (
	m "investindicator/internal/model"
	"time"

	"github.com/redis/go-redis/v9"
)

type rtPoller interface { // realtime poller
	PresentPrice(category m.Category, code string) (float64, error)
	RealEstateStatus() (string, error)
	GoldPriceDollar() (float64, error)
	AirdropEventUpbit() ([]string, []string, error)
	AirdropEventBithumb() ([]string, []string, error)
	StreamCoinOrders(c chan<- m.MyOrder) error
	StreamDomesticStockOrders(c chan<- m.MyOrder) error
}

type dailyPoller interface {
	ExchageRate() float64
	ClosingPrice(category m.Category, code string) (float64, error)
	FearGreedIndex() (uint, error)
	Nasdaq() (float64, error)
	Sp500() (float64, error)
	CliIdx() (float64, error)
	HighYieldSpread() (date string, spread float64, err error)
	RecentSP500Entries(targetDate string) ([]m.SP500Company, error)
}

type Poller interface {
	rtPoller
	dailyPoller
}

type storage interface {
	RetrieveMarketStatus(date string) (*m.Market, error)

	RetrieveAssetList() ([]m.Asset, error)
	RetrieveAsset(id uint) (*m.Asset, error)
	RetrieveTotalAssets() ([]m.Asset, error)
	UpdateAssetInfo(asset m.Asset) error
	RetrieveAssetIdByCode(code string) uint

	RetreiveFundsSummaryOrderByFundId() ([]m.InvestSummary, error)
	RetreiveFundSummaryByFundId(fundId uint) ([]m.InvestSummary, error)
	UpdateInvestSummarySum(fundId uint, assetId uint, sum float64) error
	UpdateInvestSummary(fundId uint, assetId uint, change float64, price float64) error
	RetreiveFundSummaryByAssetId(id uint) ([]m.InvestSummary, error)

	SaveInvest(fundId uint, assetId uint, price float64, count float64) error

	RetrieveMarketIndicator(date string) (*m.DailyIndex, *m.CliIndex, error)
	SaveDailyMarketIndicator(fearGreedIndex uint, nasdaq float64, sp500 float64) error
	RetrieveLatestHighYieldSpread() (*m.HighYieldSpread, error)
	SaveHighYieldSpread(hy *m.HighYieldSpread) error
	RetrieveLatestSP500Entry() (*m.SP500Company, error)

	RetreiveLatestEma(assetId uint) (*m.EmaHist, error)
	SaveEmaHist(newEma *m.EmaHist) error

	RetreiveEventIsActive(eventId uint) bool
	UpdateEventIsActive(eventId uint, isActive bool) error

	SetCache(key string, value interface{}, exp time.Duration)
	GetCache(key string) *redis.StringCmd
}

type trader interface {
	Buy(category m.Category, code string, qty uint) error
}

type bcTrader interface { // blockchain trader
	SwapUsdtUsdc(isUsdcIn bool) error
	RunBlackholeDexStrategy(reportChan chan<- string) error
}

type messenger interface {
	SendMessage(msg string)
	SendButtonsAndGetResult(prompt string, options ...string) (answer string, err error)
}
