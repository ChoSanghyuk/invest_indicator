package handler

import (
	investind "investindicator"
	m "investindicator/internal/model"
)

type FundRetriever interface {
	RetreiveFundsSummaryOrderByFundId() ([]m.InvestSummary, error)
	RetreiveFundSummaryByFundId(id uint) ([]m.InvestSummary, error)
	RetreiveFundInvestsById(id uint) ([]m.Invest, error)
	RetrieveFundInvestsByIdAndRange(id uint, start, end string) ([]m.Invest, error)
}

type FundWriter interface {
	SaveFund(name string) error
}

type AssetRetriever interface {
	RetrieveAssetList() ([]m.Asset, error)
	RetrieveAsset(id uint) (*m.Asset, error)
	RetrieveAssetHist(id uint) ([]m.Invest, error)
	RetrieveAssetIdByName(name string) uint
	RetrieveAssetIdByCode(code string) uint
}

type AssetInfoSaver interface {
	SaveAssetInfo(asset m.Asset) (uint, error)
	UpdateAssetInfo(asset m.Asset) error
	DeleteAssetInfo(id uint) error
	SaveEmaHist(newEma *m.EmaHist) error
}

type PriceGetter interface {
	TopBottomPrice(category m.Category, code string) (float64, float64, error)
	AvgPrice(category m.Category, code string) (ap float64, n uint, err error)
	PresentPrice(category m.Category, code string) (float64, error)
}

type MaketRetriever interface {
	RetrieveMarketStatus(date string) (*m.Market, error)
	RetrieveMarketIndicator(date string) (*m.DailyIndex, *m.CliIndex, error)
	RetrieveMarketIndicatorWeekDesc() ([]m.DailyIndex, error)
}

type MarketSaver interface {
	SaveMarketStatus(status uint) error
}

type InvestRetriever interface {
	RetrieveInvestHist(fundId uint, assetId uint, start string, end string) ([]m.Invest, error)
	// RetrieveInitAmountofAsset(fundId, assetId uint) (float64, error)
}

type InvestSaver interface {
	SaveInvest(fundId uint, assetId uint, price float64, count float64) error
	UpdateInvestSummary(fundId uint, assetId uint, change float64, price float64) error
}

type ExchageRateGetter interface {
	ExchageRate() float64
}

type EventRetriever interface {
	Events() []*investind.EnrolledEvent
}

type EventLauncher interface {
	LaunchEvent(id uint) error
}

type EventStatusChanger interface {
	SetEventStatus(id uint, active bool) error
}

type UserRetrierver interface {
	User(userName string) (*m.User, error)
}
