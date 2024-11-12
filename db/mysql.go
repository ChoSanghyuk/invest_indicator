package db

import (
	"database/sql"
	m "invest/model"
	"math"
	"time"

	"gorm.io/datatypes"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Storage struct {
	db *gorm.DB
}

func NewStorage(dsn string, opts ...gorm.Option) (*Storage, error) {
	sqlDB, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn: sqlDB,
	}), opts...) //&gorm.Config{}
	if err != nil {
		return nil, err
	}

	return &Storage{
		db: db,
	}, nil
}

func (s Storage) RetreiveFundsSummaryOrderByFundId() ([]m.InvestSummary, error) {

	var fundsSummary []m.InvestSummary

	result := s.db.Model(&m.InvestSummary{}).Preload("Fund").Preload("Asset").Order("fund_id").Find(&fundsSummary)

	if result.Error != nil {
		return nil, result.Error
	}

	return fundsSummary, nil

}

func (s Storage) RetreiveFundSummaryByFundId(id uint) ([]m.InvestSummary, error) {

	var fundsSummary []m.InvestSummary

	result := s.db.Model(&m.InvestSummary{}).Preload("Asset").Where("fund_id", id).Find(&fundsSummary) // .Order("asset_id")

	if result.Error != nil {
		return nil, result.Error
	}

	return fundsSummary, nil

}

func (s Storage) RetreiveFundSummaryByAssetId(id uint) ([]m.InvestSummary, error) {

	var fundsSummary []m.InvestSummary

	result := s.db.Model(&m.InvestSummary{}).Where("asset_id", id).Find(&fundsSummary) // .Order("asset_id")

	if result.Error != nil {
		return nil, result.Error
	}

	return fundsSummary, nil

}

func (s Storage) RetreiveAFundInvestsById(id uint) ([]m.Invest, error) {
	var invets []m.Invest

	result := s.db.Model(&m.Invest{}).
		Where(&m.Invest{FundID: id}, "fund_id").
		Preload("Asset").
		Find(&invets) // .Order("asset_id")

	if result.Error != nil {
		return nil, result.Error
	}

	return invets, nil
}

func (s Storage) RetreiveInvestHistOfFundById(id uint) (*m.Fund, error) {
	var fund m.Fund

	result := s.db.First(&fund, id)
	if result.Error != nil {
		return nil, result.Error
	}

	return &fund, nil
}

func (s Storage) SaveFund(name string) error {

	result := s.db.Create(&m.Fund{
		Name: name,
	})

	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (s Storage) RetrieveAssetList() ([]m.Asset, error) {

	var assets []m.Asset

	result := s.db.Model(&m.Asset{}).Select("id", "name").Find(&assets)
	if result.Error != nil {
		return nil, result.Error
	}
	return assets, nil
}

func (s Storage) RetrieveTotalAssets() ([]m.Asset, error) {
	var assets []m.Asset

	result := s.db.Model(&m.Asset{}).Find(&assets)
	if result.Error != nil {
		return nil, result.Error
	}
	return assets, nil
}

func (s Storage) RetrieveAsset(id uint) (*m.Asset, error) {

	var asset m.Asset

	result := s.db.First(&asset, id) // memo. First, Last와 같은 메소드는 대상이 없을 때 error 반환
	if result.Error != nil {
		return nil, result.Error
	}

	return &asset, nil
}

func (s Storage) RetrieveAssetHist(id uint) ([]m.Invest, error) {

	var invests []m.Invest

	result := s.db.Model(&m.Invest{}).Where("asset_id = ?", id).Preload("Asset").Find(&invests)
	if result.Error != nil {
		return nil, result.Error
	}

	return invests, nil
}

func (s Storage) RetrieveAssetIdByName(name string) uint {
	var asset m.Asset

	result := s.db.Model(&m.Asset{}).Where("name", name).Select("id").Find(&asset)
	if result.Error != nil || result.RowsAffected == 0 {
		return 0
	}

	return asset.ID
}

func (s Storage) RetrieveAssetIdByCode(code string) uint {
	var asset m.Asset

	result := s.db.Model(&m.Asset{}).Where("code", code).Select("id").Find(&asset)
	if result.Error != nil || result.RowsAffected == 0 { // memo. RowsAffected selete된 갯수 파악 가능
		return 0
	}

	return asset.ID
}

func (s Storage) SaveAssetInfo(name string, category m.Category, code string, currency string, top float64, bottom float64, selPrice float64, buyPrice float64) (uint, error) {

	asset := m.Asset{
		Name:      name,
		Category:  category,
		Code:      code,
		Currency:  currency,
		Top:       top,
		Bottom:    bottom,
		SellPrice: selPrice,
		BuyPrice:  buyPrice,
	}

	result := s.db.Create(&asset)

	if result.Error != nil {
		return 0, result.Error
	}

	return asset.ID, nil
}

// When updating with struct, GORM will only update non-zero fields. You might want to use map to update attributes or use Select to specify fields to update
func (s Storage) UpdateAssetInfo(id uint, name string, category m.Category, code string, currency string, top float64, bottom float64, selPrice float64, buyPrice float64) error {

	result := s.db.Updates(m.Asset{
		ID:        id,
		Name:      name,
		Category:  category,
		Code:      code,
		Currency:  currency,
		Top:       top,
		Bottom:    bottom,
		SellPrice: selPrice,
		BuyPrice:  buyPrice,
	})

	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (s Storage) DeleteAssetInfo(id uint) error {

	result := s.db.Delete(&m.Asset{}, id)

	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (s Storage) RetrieveMarketStatus(date string) (*m.Market, error) {

	var market m.Market

	if date == "" {
		result := s.db.Last(&market) // Preload("Asset")
		if result.Error != nil {
			return nil, result.Error
		}
	} else {
		result := s.db.Where("created_at = ?", date).Last(&market, date) // Preload("Asset")
		if result.Error != nil {
			return nil, result.Error
		}
	}

	return &market, nil
}

func (s Storage) RetrieveMarketIndicator(date string) (*m.DailyIndex, *m.CliIndex, error) {

	var dailyIdx m.DailyIndex
	var cliIdx m.CliIndex

	if date == "" {
		result := s.db.Last(&dailyIdx) // Preload("Asset")
		if result.Error != nil {
			return nil, nil, result.Error
		}

		// result = s.db.Last(&cliIdx) // Preload("Asset") // todo. CLI Index 우선 미사용
		// if result.Error != nil {
		// 	return nil, nil, result.Error
		// }
	} else {
		// memo. createdAt을 PK로 지정했더라도, First에 인자로 넣어서 where절 만들 수 없음
		result := s.db.Where("created_at = ?", date).First(&dailyIdx) // Preload("Asset")
		if result.Error != nil {
			return nil, nil, result.Error
		}

		// result = s.db.First(&cliIdx, date) // Preload("Asset")
		// if result.Error != nil {
		// 	return nil, nil, result.Error
		// }
	}

	return &dailyIdx, &cliIdx, nil
}

func (s Storage) SaveDailyMarketIndicator(fearGreedIndex uint, nasdaq float64) error {

	result := s.db.Create(&m.DailyIndex{
		CreatedAt:      datatypes.Date(time.Now()),
		FearGreedIndex: fearGreedIndex,
		NasDaq:         nasdaq,
	})
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (s Storage) SaveMarketStatus(status uint) error {

	result := s.db.Create(&m.Market{
		CreatedAt: datatypes.Date(time.Now()),
		Status:    status,
	})
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (s Storage) RetrieveInvestHist(fundId uint, assetId uint, start string, end string) ([]m.Invest, error) {

	query := s.db.Model(&m.Invest{}) // Note. 필수가 아니더라도, 처음에 모델을 명시하는 것이 good practice

	if fundId != 0 {
		query.Where("fund_id = ?", fundId)
	}
	if assetId != 0 {
		query.Where("asset_id = ?", assetId)
	}
	if start != "" {
		query.Where("created_at >= ?", start)
	}
	if end != "" {
		query.Where("created_at <= ?", end)
	}

	var investHist []m.Invest

	result := query.Preload("Asset").Find(&investHist)
	if result.Error != nil {
		return nil, result.Error
	}

	return investHist, nil
}

func (s Storage) SaveInvest(fundId uint, assetId uint, price float64, count float64) error {

	result := s.db.Create(&m.Invest{
		FundID:  fundId,
		AssetID: assetId,
		Price:   price,
		Count:   count,
	})
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (s Storage) RetrieveInvestSummaryByFundIdAssetId(fundId uint, assetId uint) (*m.InvestSummary, error) {
	var investSummary m.InvestSummary

	result := s.db.Model(&m.InvestSummary{}).
		Where("fund_id = ?", fundId).
		Where("asset_id = ?", assetId).
		First(&investSummary) // Preload("Asset")
	if result.Error != nil {
		return nil, result.Error
	}

	return &investSummary, nil
}

// memo. struct를 사용해서 Updates하는 경우에는 0이 아닌 필드만 업데이트. 이로인해 Invest 매도 기록하면서, 전체 수량이 0이 되어 업데이트 안 되는 현상 발생
// => 업데이트 필드 명시 필요
func (s Storage) UpdateInvestSummary(fundId uint, assetId uint, change float64, price float64) error {

	var investSummary m.InvestSummary
	result := s.db.Model(&m.InvestSummary{}).
		Where("fund_id = ?", fundId).
		Where("asset_id = ?", assetId).
		Find(&investSummary) // memo. Select는 필드 지정하는 용도. 조회에서 구조체에 넣으려면 Find 사용

	if result.RowsAffected == 0 {
		investSummary = m.InvestSummary{
			FundID:  fundId,
			AssetID: assetId,
			Count:   change,
			Sum:     change * price,
		}

		result = s.db.Model(&m.InvestSummary{}).Create(&investSummary)
	} else {
		investSummary.Count += change
		investSummary.Sum += change * price

		result = s.db.Model(&investSummary).Select("Count", "Sum").Updates(investSummary)
	}
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (s Storage) UpdateInvestSummarySum(fundId uint, assetId uint, sum float64) error {
	// 조회한 InvestSummary를 sum만 변경
	var investSummary m.InvestSummary

	result := s.db.Model(&m.InvestSummary{}).
		Where("fund_id = ?", fundId).
		Where("asset_id = ?", assetId).
		First(&investSummary)
	if result.Error != nil {
		return result.Error
	}

	s.db.Model(&investSummary).Update("sum", sum)
	return nil
}

func (s Storage) RetreiveLatestEma(assetId uint) (float64, error) {

	var ema m.EmaHist
	// result := s.db.Where("asset_id", assetId).Order("date desc").First(&ema)
	result := s.db.Where("asset_id", assetId).Last(&ema)
	if result.Error != nil {
		return 0, result.Error
	}

	return ema.Ema, nil
}

func (s Storage) SaveEmaHist(assetId uint, price float64) error {

	emay, err := s.RetreiveLatestEma(assetId)
	if err != nil {
		emay = price
	}

	var ema = m.EmaHist{
		AssetID: assetId,
		Date:    datatypes.Date(time.Now()),
		Ema:     ema(price, emay),
	}

	result := s.db.Create(&ema)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

/*
a =  2/N+1
EMAt = a*PRICEt + (1-a)EMAy
*/
func ema(tp float64, emay float64) float64 {

	a := 2.0 / (200 + 1)
	return math.Round((a*tp+(1-a)*emay)*100) / 100
}
