package event

import (
	m "invest/model"
)

type Storage interface {
	RetrieveMarketStatus(date string) (*m.Market, error)

	RetrieveAssetList() ([]m.Asset, error)
	RetrieveAsset(id uint) (*m.Asset, error)
	RetrieveTotalAssets() ([]m.Asset, error)
	UpdateAssetInfo(asset m.Asset) error

	RetreiveFundsSummaryOrderByFundId() ([]m.InvestSummary, error)
	UpdateInvestSummarySum(fundId uint, assetId uint, sum float64) error
	RetreiveFundSummaryByAssetId(id uint) ([]m.InvestSummary, error)

	RetrieveMarketIndicator(date string) (*m.DailyIndex, *m.CliIndex, error)
	SaveDailyMarketIndicator(fearGreedIndex uint, nasdaq float64, sp500 float64) error

	RetreiveLatestEma(assetId uint) (*m.EmaHist, error)
	SaveEmaHist(newEma *m.EmaHist) error

	RetreiveEventIsActive(eventId uint) bool
	UpdateEventIsActive(eventId uint, isActive bool) error
}

type RtPoller interface {
	PresentPrice(category m.Category, code string) (float64, error)
	RealEstateStatus() (string, error)
	GoldPriceDollar() (float64, error)
}

type DailyPoller interface {
	ExchageRate() float64
	ClosingPrice(category m.Category, code string) (float64, error)
	FearGreedIndex() (uint, error)
	Nasdaq() (float64, error)
	Sp500() (float64, error)
	CliIdx() (float64, error)
}
