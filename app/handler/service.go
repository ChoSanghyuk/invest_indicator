package handler

import (
	m "invest/model"
)

type FundRetriever interface {
	RetreiveFundsSummaryOrderByFundId() ([]m.InvestSummary, error)
	RetreiveFundSummaryByFundId(id uint) ([]m.InvestSummary, error)
	RetreiveAFundInvestsById(id uint) ([]m.Invest, error)
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
	SaveAssetInfo(name string, category m.Category, code string, currency string, top float64, bottom float64, selPrice float64, buyPrice float64) (uint, error)
	UpdateAssetInfo(id uint, name string, category m.Category, code string, currency string, top float64, bottom float64, selPrice float64, buyPrice float64) error
	DeleteAssetInfo(id uint) error
	SaveEmaHist(assetId uint, price float64) error
}

type TopBottomPriceGetter interface {
	TopBottomPrice(category m.Category, code string) (float64, float64, error)
}

type MaketRetriever interface {
	RetrieveMarketStatus(date string) (*m.Market, error)
	RetrieveMarketIndicator(date string) (*m.DailyIndex, *m.CliIndex, error)
	RetrieveMarketIndicatorWeek() ([]m.DailyIndex, error)
}

type MarketSaver interface {
	SaveMarketStatus(status uint) error
}

type InvestRetriever interface {
	RetrieveInvestHist(fundId uint, assetId uint, start string, end string) ([]m.Invest, error)
}

type InvestSaver interface {
	SaveInvest(fundId uint, assetId uint, price float64, count float64) error
	UpdateInvestSummary(fundId uint, assetId uint, change float64, price float64) error
}

type ExchageRateGetter interface {
	ExchageRate() float64
}
