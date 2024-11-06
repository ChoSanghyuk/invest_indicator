package handler

import (
	"fmt"
	m "invest/model"
	"time"

	"gorm.io/datatypes"
)

/***************************** Asset ***********************************/
type AssetRetrieverMock struct {
	err error
}

func (mock AssetRetrieverMock) RetrieveAssetList() ([]m.Asset, error) {
	fmt.Println("RetrieveAssetList Called")

	if mock.err != nil {
		return nil, mock.err
	}
	return []m.Asset{
		{
			ID:   1,
			Name: "비트코인",
		},
		{
			ID:   2,
			Name: "TigerS&P500",
		},
	}, nil
}
func (mock AssetRetrieverMock) RetrieveAsset(id uint) (*m.Asset, error) {
	fmt.Println("RetrieveAsset Called")

	if mock.err != nil {
		return nil, mock.err
	}
	return &m.Asset{
		ID:        1,
		Name:      "bitcoin",
		Code:      "BTS",
		Category:  m.ForeignStock,
		Currency:  "USD",
		Top:       9800,
		Bottom:    6800,
		SellPrice: 8800,
		BuyPrice:  7800,
	}, nil
}

func (mock AssetRetrieverMock) RetrieveAssetHist(id uint) ([]m.Invest, error) {
	fmt.Println("RetrieveAssetHist Called")

	if mock.err != nil {
		return nil, mock.err
	}
	return []m.Invest{
		{
			ID:      1,
			FundID:  3,
			AssetID: 1,
			Price:   7800,
			Count:   5,
		},
	}, nil
}

func (mock AssetRetrieverMock) RetrieveAssetIdByName(name string) uint {
	return 1
}
func (mock AssetRetrieverMock) RetrieveAssetIdByCode(code string) uint {
	return 1
}

type AssetInfoSaverMock struct {
	err error
}

func (mock AssetInfoSaverMock) SaveAssetInfo(name string, category m.Category, code string, currency string, top float64, bottom float64, selPrice float64, buyPrice float64) (uint, error) {
	fmt.Println("SaveAssetInfo Called")

	if mock.err != nil {
		return 0, mock.err
	}
	return 0, nil
}
func (mock AssetInfoSaverMock) UpdateAssetInfo(ID uint, name string, category m.Category, code string, currency string, top float64, bottom float64, selPrice float64, buyPrice float64) error {
	fmt.Println("UpdateAssetInfo Called")

	if mock.err != nil {
		return mock.err
	}
	return nil
}
func (mock AssetInfoSaverMock) DeleteAssetInfo(id uint) error {
	fmt.Println("DeleteAssetInfo Called")

	if mock.err != nil {
		return mock.err
	}
	return nil
}

func (mock AssetInfoSaverMock) SaveEmaHist(assetId uint, price float64) error {
	if mock.err != nil {
		return mock.err
	}
	return nil
}

type TopBottomPriceGetterMock struct {
	err error
}

func (mock TopBottomPriceGetterMock) TopBottomPrice(category m.Category, code string) (float64, float64, error) {
	fmt.Println("TopBottomPrice Called")

	if mock.err != nil {
		return 0, 0, mock.err
	}
	return 1000, 100, nil
}

/***************************** Fund ***********************************/
type FundRetrieverMock struct {
	isli []m.InvestSummary
	il   []m.Invest
	err  error
}

func (mock FundRetrieverMock) RetreiveFundsSummaryOrderByFundId() ([]m.InvestSummary, error) {
	fmt.Println("RetreiveFundsSummary Called")

	if mock.err != nil {
		return nil, mock.err
	}
	return mock.isli, nil
}

func (mock FundRetrieverMock) RetreiveFundSummaryByFundId(id uint) ([]m.InvestSummary, error) {
	fmt.Println("RetreiveFundSummaryByFundId Called")

	if mock.err != nil {
		return nil, mock.err
	}
	return mock.isli, nil
}

func (mock FundRetrieverMock) RetreiveAFundInvestsById(id uint) ([]m.Invest, error) {
	fmt.Println("RetreiveAFundInvestsById Called")

	if mock.err != nil {
		return nil, mock.err
	}
	var rtn []m.Invest
	for _, iv := range mock.il {
		if iv.FundID == id {
			rtn = append(rtn, iv)
		}
	}
	return rtn, nil
}

type FundWriterMock struct {
	err error
}

func (mock FundWriterMock) SaveFund(name string) error {
	fmt.Println("SaveFund Called")

	if mock.err != nil {
		return mock.err
	}
	return nil
}

type ExchageRateGetterMock struct {
}

func (mock ExchageRateGetterMock) ExchageRate() float64 {
	fmt.Println("SaveFund Called")

	return 1334.3
}

/***************************** Market ***********************************/
type MaketRetrieverMock struct {
	err error
}

func (mock MaketRetrieverMock) RetrieveMarketStatus(date string) (*m.Market, error) {
	fmt.Println("RetrieveMarketStatus Called")

	if mock.err != nil {
		return nil, mock.err
	}
	return &m.Market{
		CreatedAt: datatypes.Date(time.Now()),
		Status:    3,
	}, nil
}

func (mock MaketRetrieverMock) RetrieveMarketIndicator(date string) (*m.DailyIndex, *m.CliIndex, error) {
	fmt.Println("RetrieveMarketIndicator Called")

	if mock.err != nil {
		return nil, nil, mock.err
	}
	return &m.DailyIndex{
			CreatedAt:      datatypes.Date(time.Now()),
			FearGreedIndex: 23,
			NasDaq:         17556.03,
		}, &m.CliIndex{
			CreatedAt: datatypes.Date(time.Now()),
			Index:     102,
		}, nil
}

type MarketSaverMock struct {
	err error
}

func (mock MarketSaverMock) SaveMarketStatus(status uint) error {
	fmt.Println("SaveMarketStatus Called")

	if mock.err != nil {
		return mock.err
	}
	return nil
}

/***************************** Invest ***********************************/
type InvestRetrieverMock struct {
	err error
}

func (mock InvestRetrieverMock) RetrieveInvestHist(fundId uint, assetId uint, start string, end string) ([]m.Invest, error) {
	fmt.Println("RetrieveInvestHist Called")

	if mock.err != nil {
		return nil, mock.err
	}
	return []m.Invest{
		{
			ID:      1,
			FundID:  fundId,
			AssetID: assetId,
			Price:   7800,
			Count:   5,
		},
	}, nil
}

type InvestSaverMock struct {
	err error
}

func (mock InvestSaverMock) SaveInvest(fundId uint, assetId uint, price float64, count float64) error {
	fmt.Println("SaveInvest Called")

	if mock.err != nil {
		return mock.err
	}
	return nil
}

func (mock InvestSaverMock) UpdateInvestSummary(fundId uint, assetId uint, change float64, price float64) error {
	fmt.Println("UpdateInvestSummaryCount Called")

	if mock.err != nil {
		return mock.err
	}
	return nil
}
