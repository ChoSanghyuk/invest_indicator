package handler

import (
	"fmt"
	m "investindicator/internal/model"
	"strings"
	"time"

	"github.com/kr/pretty"
)

/***************************** Asset ***********************************/

type PriceGetterMock struct {
	err error
}

func (mock PriceGetterMock) TopBottomPrice(category m.Category, code string) (float64, float64, error) {
	fmt.Println("TopBottomPrice Called")

	if mock.err != nil {
		return 0, 0, mock.err
	}
	return 1000, 100, nil
}

func (mock PriceGetterMock) AvgPrice(category m.Category, code string) (ap float64, n uint, err error) {
	return 0, 0, nil
}

func (mock PriceGetterMock) PresentPrice(category m.Category, code string) (float64, error) {
	return 0, nil
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

func (mock FundRetrieverMock) RetreiveFundInvestsById(id uint) ([]m.Invest, error) {
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
func (mock FundRetrieverMock) RetrieveFundInvestsByIdAndRange(id uint, start, end string) ([]m.Invest, error) {
	if mock.err != nil {
		return nil, mock.err
	}

	var rtn []m.Invest
	for _, iv := range mock.il {
		if iv.FundID == id &&
			strings.Compare(iv.CreatedAt.Format("2006-01-02"), start) == 1 &&
			strings.Compare(iv.CreatedAt.Format("2006-01-02"), end) == -1 {
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

/***************************** Market ***********************************/
type MaketRetrieverMock struct {
	err          error
	marketStatus uint
}

func (mock MaketRetrieverMock) RetrieveMarketStatus(date string) (*m.Market, error) {
	fmt.Println("RetrieveMarketStatus Called")

	if mock.err != nil {
		return nil, mock.err
	}

	status := mock.marketStatus
	if status == 0 {
		status = 3 // Default to VOLATILIY if not set
	}

	return &m.Market{
		CreatedAt: time.Now(),
		Status:    status,
	}, nil
}

func (mock MaketRetrieverMock) RetrieveMarketIndicator(date string) (*m.DailyIndex, *m.CliIndex, error) {
	fmt.Println("RetrieveMarketIndicator Called")

	if mock.err != nil {
		return nil, nil, mock.err
	}
	return &m.DailyIndex{
			CreatedAt:      time.Now(),
			FearGreedIndex: 23,
			NasDaq:         17556.03,
		}, &m.CliIndex{
			CreatedAt: time.Now(),
			Index:     102,
		}, nil
}

func (mock MaketRetrieverMock) RetrieveMarketIndicatorWeekDesc() ([]m.DailyIndex, error) {
	return nil, nil
}

func (mock MaketRetrieverMock) RetrieveHighYieldSpreadWeekDesc() ([]m.HighYieldSpread, error) {
	return nil, nil
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
	invests []m.Invest
	err     error
}

func (mock InvestRetrieverMock) RetrieveInvestHist(fundId uint, assetId uint, start string, end string) ([]m.Invest, error) {
	fmt.Println("RetrieveInvestHist Called")

	if mock.err != nil {
		return nil, mock.err
	}

	if mock.invests != nil {
		rtn := []m.Invest{}
		for _, iv := range mock.invests {
			if iv.FundID == fundId && iv.AssetID == assetId {
				rtn = append(rtn, iv)
			}
		}
		return rtn, nil
	} else {
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
}

func (mock InvestRetrieverMock) RetrieveInitAmountofAsset(fundId, assetId uint) (float64, error) {
	return 0, nil
}

/***************************** 작업 완료 ***********************************/
type AssetRetrieverMock struct {
	assets []m.Asset
	hist   []m.Invest
	err    error
}

func NewADefaultssetRetrieverMock(assets ...m.Asset) *AssetRetrieverMock {
	mock := &AssetRetrieverMock{
		assets: []m.Asset{
			{ID: 1, Name: "KRW", Category: m.Won, Currency: "WON", Top: 0, Bottom: 0, SellPrice: 0, BuyPrice: 0},
			{ID: 2, Name: "USD", Category: m.Dollar, Currency: "WON", Top: 0, Bottom: 0, SellPrice: 0, BuyPrice: 0},
			{ID: 3, Name: "삼성전자", Category: m.DomesticStock, Currency: "WON", Top: 9800, Bottom: 6800, SellPrice: 8800, BuyPrice: 7800},
			{ID: 4, Name: "애플", Category: m.ForeignStock, Currency: "USD", Top: 9800, Bottom: 6800, SellPrice: 8800, BuyPrice: 7800},
		}}
	mock.assets = append(mock.assets, assets...)
	return mock
}

func (mock AssetRetrieverMock) RetrieveAssetList() ([]m.Asset, error) {
	fmt.Println("RetrieveAssetList Called")

	if mock.err != nil {
		return nil, mock.err
	}
	return mock.assets, nil
}
func (mock AssetRetrieverMock) RetrieveAsset(id uint) (*m.Asset, error) {

	if mock.err != nil {
		return nil, mock.err
	}
	for _, a := range mock.assets {
		if a.ID == id {
			return &a, nil
		}
	}
	return nil, fmt.Errorf("asset not found")
}

func (mock AssetRetrieverMock) RetrieveAssetHist(id uint) ([]m.Invest, error) {
	fmt.Println("RetrieveAssetHist Called")

	if mock.err != nil {
		return nil, mock.err
	}
	return mock.hist, nil
}

func (mock AssetRetrieverMock) RetrieveAssetIdByName(name string) uint {
	for _, a := range mock.assets {
		if a.Name == name {
			return a.ID
		}
	}
	return 0
}
func (mock AssetRetrieverMock) RetrieveAssetIdByCode(code string) uint {
	for _, a := range mock.assets {
		if a.Code == code {
			return a.ID
		}
	}
	return 0
}

func (mock AssetRetrieverMock) RetreiveLatestEma(assetId uint) (*m.EmaHist, error) {
	return nil, nil
}

type AssetInfoSaverMock struct {
	assets []m.Asset
	hist   []m.EmaHist
	err    error
}

func (mock *AssetInfoSaverMock) SaveAssetInfo(a m.Asset) (uint, error) {
	if mock.err != nil {
		return 0, mock.err
	}
	a.ID = uint(len(mock.assets)) + 1
	mock.assets = append(mock.assets, a)

	return a.ID, nil
}
func (mock *AssetInfoSaverMock) UpdateAssetInfo(a m.Asset) error {
	if mock.err != nil {
		return mock.err
	}
	for i, asset := range mock.assets {
		if asset.ID == a.ID {
			mock.assets[i] = a
			return nil
		}
	}
	return fmt.Errorf("asset not found")
}
func (mock *AssetInfoSaverMock) DeleteAssetInfo(id uint) error {
	if mock.err != nil {
		return mock.err
	}

	for i, asset := range mock.assets {
		if asset.ID == id {
			mock.assets = append(mock.assets[:i], mock.assets[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("asset not found")
}

func (mock *AssetInfoSaverMock) SaveEmaHist(newEma *m.EmaHist) error {
	if mock.err != nil {
		return mock.err
	}

	newEma.ID = uint(len(mock.hist)) + 1
	mock.hist = append(mock.hist, *newEma)
	return nil
}

type ExchageRateGetterMock struct {
	exchangeRate float64
}

func NewExchageRateGetterMock(exchangeRate float64) *ExchageRateGetterMock {
	return &ExchageRateGetterMock{exchangeRate: exchangeRate}
}

func (mock ExchageRateGetterMock) ExchageRate() float64 {

	return mock.exchangeRate
}

type InvestSaverMock struct {
	initAmount float64
	invests    []m.Invest
	summaries  []m.InvestSummary
	err        error
}

func NewInvestSaverMock(initAmount float64) *InvestSaverMock {
	return &InvestSaverMock{
		initAmount: initAmount,
		invests: []m.Invest{
			{ID: 1, FundID: 1, AssetID: 1, Price: 1, Count: initAmount},
		},
		summaries: []m.InvestSummary{
			{ID: 1, FundID: 1, AssetID: 1, Sum: initAmount, Count: initAmount},
		},
	}
}

func (mock *InvestSaverMock) reset() {
	mock.invests = []m.Invest{
		{ID: 1, FundID: 1, AssetID: 1, Price: 1, Count: mock.initAmount},
	}
	mock.summaries = []m.InvestSummary{
		{ID: 1, FundID: 1, AssetID: 1, Sum: mock.initAmount, Count: mock.initAmount},
	}
}

func (mock *InvestSaverMock) prettyPrint() {

	pretty.Println(mock.invests)
	pretty.Println(mock.summaries)
}

func (mock *InvestSaverMock) SaveInvest(fundId uint, assetId uint, price float64, count float64) error {
	if mock.err != nil {
		return mock.err
	}
	mock.invests = append(mock.invests, m.Invest{
		FundID:  fundId,
		AssetID: assetId,
		Price:   price,
		Count:   count,
	})
	return nil
}

func (mock *InvestSaverMock) UpdateInvestSummary(fundId uint, assetId uint, change float64, price float64) error {
	if mock.err != nil {
		return mock.err
	}

	isExist := false
	for i, s := range mock.summaries {
		if s.FundID == fundId && s.AssetID == assetId {
			mock.summaries[i].Count += change
			mock.summaries[i].Sum += change * price
			isExist = true
			break
		}
	}

	if !isExist {
		mock.summaries = append(mock.summaries, m.InvestSummary{
			FundID:  fundId,
			AssetID: assetId,
			Count:   change,
			Sum:     change * price,
		})
	}
	return nil
}

/***************************** InvestStatusIndicator ***********************************/
type InvestStatusIndicatorMock struct {
	availableAmount float64
	err             error
}

func NewInvestStatusIndicatorMock(availableAmount float64) *InvestStatusIndicatorMock {
	return &InvestStatusIndicatorMock{
		availableAmount: availableAmount,
	}
}

func (mock InvestStatusIndicatorMock) InvestAvailableAmount(fundId int) (float64, error) {
	fmt.Printf("InvestAvailableAmount Called with fundId: %d\n", fundId)

	if mock.err != nil {
		return 0, mock.err
	}

	return mock.availableAmount, nil
}
