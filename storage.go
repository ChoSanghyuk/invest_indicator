package investind

import (
	m "investindicator/internal/model"
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
	RetrieveLatestHighYieldSpread() (*m.HighYieldSpread, error)
	SaveHighYieldSpread(hy *m.HighYieldSpread) error

	RetreiveLatestEma(assetId uint) (*m.EmaHist, error)
	SaveEmaHist(newEma *m.EmaHist) error

	RetreiveEventIsActive(eventId uint) bool
	UpdateEventIsActive(eventId uint, isActive bool) error
}
