package app

import m "invest/model"

type Storage interface {
	RetreiveFundsSummaryOrderByFundId() ([]m.InvestSummary, error)
	RetreiveFundSummaryByFundId(id uint) ([]m.InvestSummary, error)
	RetreiveFundSummaryByAssetId(id uint) ([]m.InvestSummary, error)
	RetreiveFundInvestsById(id uint) ([]m.Invest, error)
	RetrieveFundInvestsByIdAndRange(fundID uint, startDate, endDate string) ([]m.Invest, error)
	SaveFund(name string) error
	RetrieveAssetList() ([]m.Asset, error)
	RetrieveTotalAssets() ([]m.Asset, error)
	RetrieveAsset(id uint) (*m.Asset, error)
	RetrieveAssetHist(id uint) ([]m.Invest, error)
	RetrieveAssetIdByName(name string) uint
	RetrieveAssetIdByCode(code string) uint
	SaveAssetInfo(asset m.Asset) (uint, error)
	UpdateAssetInfo(asset m.Asset) error
	DeleteAssetInfo(id uint) error
	RetrieveMarketStatus(date string) (*m.Market, error)
	RetrieveMarketIndicator(date string) (*m.DailyIndex, *m.CliIndex, error)
	RetrieveMarketIndicatorWeekDesc() ([]m.DailyIndex, error)
	SaveDailyMarketIndicator(fearGreedIndex uint, nasdaq float64, sp500 float64) error
	SaveMarketStatus(status uint) error
	RetrieveInvestHist(fundId uint, assetId uint, start string, end string) ([]m.Invest, error)
	SaveInvest(fundId uint, assetId uint, price float64, count float64) error
	RetrieveInvestSummaryByFundIdAssetId(fundId uint, assetId uint) (*m.InvestSummary, error)
	UpdateInvestSummary(fundId uint, assetId uint, change float64, price float64) error
	UpdateInvestSummarySum(fundId uint, assetId uint, sum float64) error
	RetreiveLatestEma(assetId uint) (*m.EmaHist, error)
	SaveEmaHist(newEma *m.EmaHist) error
	User(userName string) (*m.User, error)
	RetreiveEventIsActive(eventId uint) bool
	UpdateEventIsActive(eventId uint, isActive bool) error
}

type Scraper interface {
	PresentPrice(category m.Category, code string) (pp float64, err error)
	TopBottomPrice(category m.Category, code string) (hp float64, lp float64, err error)
	AvgPrice(category m.Category, code string) (float64, uint, error)
	ClosingPrice(category m.Category, code string) (cp float64, err error)
	RealEstateStatus() (string, error)
	ExchageRate() float64
	FearGreedIndex() (uint, error)
	Nasdaq() (float64, error)
	Sp500() (float64, error)
	CliIdx() (float64, error)
	GoldPriceDollar() (float64, error)
	Buy(category m.Category, code string) error
}

type EventHandler interface {
	Events() []*EnrolledEvent
	StatusChange(id uint, active bool) error
	Launch(id uint) error
	AssetEvent()
	CoinEvent()
	AssetRecommendEvent(isManual WayOfLaunch)
	coinKimchiPremiumEvent(isManual WayOfLaunch)
	EmaUpdateEvent()
	RealEstateEvent()
	IndexEvent()
	assetUpdate(priceMap map[uint]float64, ivsmLi *[]m.InvestSummary)
	buySellMsg(assetId uint, pm map[uint]float64) (msg string, err error)
	hasIt(id uint) bool
	updateFundSummarys(list []m.InvestSummary, pm map[uint]float64) (err error)
	portfolioMsg(ivsmLi []m.InvestSummary, pm map[uint]float64) (msg string, err error)
	loadOrderSlice(os *[]priority, pm map[uint]float64) error
	goldKimchiPremium(isManual WayOfLaunch)
}
