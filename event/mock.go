package event

import (
	m "invest/model"
	md "invest/model"
)

type StorageMock struct {
	ma     map[uint]float64
	market *md.Market
	assets []md.Asset
	ivsm   []md.InvestSummary
	err    error
}

func (m StorageMock) RetrieveMarketStatus(date string) (*md.Market, error) {
	if m.err != nil {
		return nil, m.err
	}

	return m.market, nil
}

func (m StorageMock) RetrieveAssetList() ([]md.Asset, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.assets, nil
}

func (m StorageMock) RetrieveAsset(id uint) (*md.Asset, error) {
	if m.err != nil {
		return nil, m.err
	}
	for _, a := range m.assets {
		if a.ID == id {
			return &a, nil
		}
	}
	return &md.Asset{}, nil
}

func (m StorageMock) UpdateAssetInfo(id uint, name string, category md.Category, code string, currency string, top float64, bottom float64, selPrice float64, buyPrice float64) error {
	return nil
}

func (m StorageMock) RetreiveFundsSummaryOrderByFundId() ([]md.InvestSummary, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.ivsm, nil
}

func (m StorageMock) UpdateInvestSummarySum(fundId uint, assetId uint, sum float64) error {
	if m.err != nil {
		return m.err
	}
	return nil
}

// todo. 목 수정
func (m StorageMock) RetrieveMarketIndicator(date string) (*md.DailyIndex, *md.CliIndex, error) {
	return nil, nil, nil
}

func (m StorageMock) SaveDailyMarketIndicator(fearGreedIndex uint, nasdaq float64) error {
	if m.err != nil {
		return m.err
	}
	return nil
}

func (m StorageMock) RetreiveLatestEma(assetId uint) (float64, error) {
	return m.ma[assetId], nil
}
func (m StorageMock) SaveEmaHist(assetId uint, price float64) error {
	return nil
}

func (m StorageMock) RetrieveTotalAssets() ([]md.Asset, error) {
	return nil, nil
}

func (m StorageMock) RetreiveFundSummaryByAssetId(id uint) ([]m.InvestSummary, error) {
	return nil, nil
}

type RtPollerMock struct {
	pp     float64
	estate string
	err    error
}

func (m RtPollerMock) PresentPrice(category md.Category, code string) (float64, error) {
	if m.err != nil {
		return 0, m.err
	}
	return m.pp, nil
}

func (m RtPollerMock) RealEstateStatus() (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.estate, nil
}

type DailyPollerMock struct {
	err error
}

func (m DailyPollerMock) ExchageRate() float64 {
	if m.err != nil {
		return 0
	}

	return 1300
}

func (m DailyPollerMock) FearGreedIndex() (uint, error) {
	return 0, nil
}
func (m DailyPollerMock) Nasdaq() (float64, error) {
	return 0, nil
}
func (m DailyPollerMock) CliIdx() (float64, error) {
	return 0, nil
}

func (m DailyPollerMock) ClosingPrice(category md.Category, code string) (float64, error) {
	return 0, nil
}
