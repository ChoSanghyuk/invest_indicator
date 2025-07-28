package investind

import (
	md "investindicator/internal/model"
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

func (m StorageMock) UpdateAssetInfo(md.Asset) error {
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

func (m StorageMock) SaveDailyMarketIndicator(fearGreedIndex uint, nasdaq float64, sp500 float64) error {
	if m.err != nil {
		return m.err
	}
	return nil
}

func (m StorageMock) RetreiveLatestEma(assetId uint) (*md.EmaHist, error) {
	return &md.EmaHist{
		Ema: m.ma[assetId],
	}, nil
}

func (m StorageMock) SaveEmaHist(newEma *md.EmaHist) error {
	return nil
}

func (m StorageMock) RetrieveTotalAssets() ([]md.Asset, error) {
	return m.assets, nil
}

func (m StorageMock) RetreiveFundSummaryByAssetId(id uint) ([]md.InvestSummary, error) {
	return nil, nil
}

func (m StorageMock) RetreiveEventIsActive(eventId uint) bool {
	return false
}

func (m StorageMock) UpdateEventIsActive(eventId uint, isActive bool) error {
	return nil
}

func (m StorageMock) RetrieveLatestHighYieldSpread() (*md.HighYieldSpread, error) {
	return nil, nil
}

func (m StorageMock) SaveHighYieldSpread(hy *md.HighYieldSpread) error {
	return nil
}

func (m StorageMock) RetrieveLatestSP500Entry() (*md.SP500Company, error) {
	return nil, nil
}
