package event

import (
	m "invest/model"
)

type Storage interface {
	RetrieveMarketStatus(date string) (*m.Market, error)

	RetrieveAssetList() ([]m.Asset, error)
	RetrieveAsset(id uint) (*m.Asset, error)
	RetrieveTotalAssets() ([]m.Asset, error)
	UpdateAssetInfo(id uint, name string, category m.Category, code string, currency string, top float64, bottom float64, selPrice float64, buyPrice float64) error

	RetreiveFundsSummaryOrderByFundId() ([]m.InvestSummary, error)
	UpdateInvestSummarySum(fundId uint, assetId uint, sum float64) error
	RetreiveFundSummaryByAssetId(id uint) ([]m.InvestSummary, error)

	RetrieveMarketIndicator(date string) (*m.DailyIndex, *m.CliIndex, error)
	SaveDailyMarketIndicator(fearGreedIndex uint, nasdaq float64) error

	RetreiveLatestEma(assetId uint) (float64, error)
	SaveEmaHist(assetId uint, price float64) error
}

type RtPoller interface {
	PresentPrice(category m.Category, code string) (float64, error)
	RealEstateStatus() (string, error)
}

type DailyPoller interface {
	ExchageRate() float64
	ClosingPrice(category m.Category, code string) (float64, error)
	FearGreedIndex() (uint, error)
	Nasdaq() (float64, error)
	CliIdx() (float64, error)
}
