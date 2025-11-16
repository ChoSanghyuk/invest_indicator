package db

import (
	"errors"
	"fmt"
	m "investindicator/internal/model"
	"time"

	"gorm.io/gorm"
)

func stgDsn(conf *MysqlConfig) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", conf.user, conf.password, conf.ip, conf.port, conf.scheme)
}

func (s Storage) initTables() error {

	err := s.db.AutoMigrate(&m.Fund{}, &m.Asset{}, &m.EmaHist{},
		&m.Invest{}, &m.InvestSummary{}, &m.Market{},
		&m.DailyIndex{}, &m.CliIndex{}, &m.HighYieldSpread{},
		&m.User{}, &m.Event{}, &m.SP500Company{})
	if err != nil {
		panic("failed to migrate database")
	}
	return nil
}

func (s Storage) RetreiveFundsSummaryOrderByFundId() ([]m.InvestSummary, error) {

	var fundsSummary []m.InvestSummary

	result := s.db.Model(&m.InvestSummary{}).Preload("Fund").Preload("Asset").Order("fund_id").Find(&fundsSummary)

	if result.Error != nil {
		return nil, result.Error
	}

	s.lg.Info().Msgf("Retrieved %d funds summary ordered by fund ID", len(fundsSummary))
	return fundsSummary, nil

}

func (s Storage) RetreiveFundSummaryByFundId(id uint) ([]m.InvestSummary, error) {

	var fundsSummary []m.InvestSummary

	result := s.db.Model(&m.InvestSummary{}).Preload("Asset").Where("fund_id", id).Find(&fundsSummary) // .Order("asset_id")

	if result.Error != nil {
		return nil, result.Error
	}

	s.lg.Info().Msgf("Retrieved %d fund summaries for fund ID %d", len(fundsSummary), id)
	return fundsSummary, nil

}

func (s Storage) RetreiveFundSummaryByAssetId(id uint) ([]m.InvestSummary, error) {

	var fundsSummary []m.InvestSummary

	result := s.db.Model(&m.InvestSummary{}).Where("asset_id", id).Find(&fundsSummary) // .Order("asset_id")

	if result.Error != nil {
		return nil, result.Error
	}

	s.lg.Info().Msgf("Retrieved %d fund summaries for asset ID %d", len(fundsSummary), id)
	return fundsSummary, nil

}

func (s Storage) RetreiveFundInvestsById(id uint) ([]m.Invest, error) {
	var invets []m.Invest

	result := s.db.Model(&m.Invest{}).
		Where(&m.Invest{FundID: id}, "fund_id").
		Preload("Asset").
		Find(&invets) // .Order("asset_id")

	if result.Error != nil {
		return nil, result.Error
	}

	s.lg.Info().Msgf("Retrieved %d invests for fund ID %d", len(invets), id)
	return invets, nil
}

func (s Storage) RetrieveFundInvestsByIdAndRange(fundID uint, startDate, endDate string) ([]m.Invest, error) {
	var invests []m.Invest

	from, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, err
	}

	temp, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil, err
	}
	end := time.Date(
		temp.Year(), temp.Month(), temp.Day(),
		23, 59, 59, 999,
		temp.Location(),
	)

	err = s.db.Where("fund_id = ? AND created_at BETWEEN ? AND ?", fundID, from, end).
		Preload("Asset").
		Find(&invests).
		Error

	s.lg.Info().Msgf("Retrieved %d invests for fund ID %d between %s and %s", len(invests), fundID, startDate, endDate)
	return invests, err
}

// func (s Storage) RetreiveInvestHistOfFundById(id uint) (*m.Fund, error) {
// 	var fund m.Fund

// 	result := s.db.First(&fund, id)
// 	if result.Error != nil {
// 		return nil, result.Error
// 	}

// 	return &fund, nil
// }

func (s Storage) SaveFund(name string) error {

	result := s.db.Create(&m.Fund{
		Name: name,
	})

	if result.Error != nil {
		return result.Error
	}

	s.lg.Info().Msgf("Saved fund with name %s", name)
	return nil
}

func (s Storage) RetrieveAssetList() ([]m.Asset, error) {

	var assets []m.Asset

	result := s.db.Model(&m.Asset{}).Find(&assets)
	if result.Error != nil {
		return nil, result.Error
	}

	s.lg.Info().Msgf("Retrieved %d assets", len(assets))
	return assets, nil
}

func (s Storage) RetrieveTotalAssets() ([]m.Asset, error) {
	var assets []m.Asset

	result := s.db.Model(&m.Asset{}).Find(&assets)
	if result.Error != nil {
		return nil, result.Error
	}

	s.lg.Info().Msgf("Retrieved total of %d assets", len(assets))
	return assets, nil
}

func (s Storage) RetrieveAsset(id uint) (*m.Asset, error) {

	var asset m.Asset

	result := s.db.First(&asset, id) // memo. First, Last와 같은 메소드는 대상이 없을 때 error 반환
	if result.Error != nil {
		return nil, result.Error
	}

	s.lg.Info().Msgf("Retrieved asset with ID %d", id)
	return &asset, nil
}

func (s Storage) RetrieveAssetHist(id uint) ([]m.Invest, error) {

	var invests []m.Invest

	result := s.db.Model(&m.Invest{}).Where("asset_id = ?", id).Preload("Asset").Find(&invests)
	if result.Error != nil {
		return nil, result.Error
	}

	s.lg.Info().Msgf("Retrieved %d invests for asset ID %d", len(invests), id)
	return invests, nil
}

func (s Storage) RetrieveAssetIdByName(name string) uint {
	var asset m.Asset

	result := s.db.Model(&m.Asset{}).Where("name", name).Select("id").Find(&asset)
	if result.Error != nil || result.RowsAffected == 0 {
		s.lg.Info().Msgf("No asset found with name %s", name)
		return 0
	}

	s.lg.Info().Msgf("Retrieved asset ID %d for name %s", asset.ID, name)
	return asset.ID
}

func (s Storage) RetrieveAssetIdByCode(code string) uint {
	var asset m.Asset

	result := s.db.Model(&m.Asset{}).Where("code", code).Select("id").Find(&asset)
	if result.Error != nil || result.RowsAffected == 0 { // memo. RowsAffected selete된 갯수 파악 가능
		s.lg.Info().Msgf("No asset found with code %s", code)
		return 0
	}

	s.lg.Info().Msgf("Retrieved asset ID %d for code %s", asset.ID, code)
	return asset.ID
}

// todo. currency 조정 필요해 보임
func (s Storage) SaveAssetInfo(asset m.Asset) (uint, error) {

	// asset := m.Asset{
	// 	Name:      name,
	// 	Category:  category,
	// 	Code:      code,
	// 	Currency:  currency,
	// 	Top:       top,
	// 	Bottom:    bottom,
	// 	SellPrice: selPrice,
	// 	BuyPrice:  buyPrice,
	// }

	result := s.db.Create(&asset)

	if result.Error != nil {
		return 0, result.Error
	}

	s.lg.Info().Msgf("Saved asset with ID %d", asset.ID)
	return asset.ID, nil
}

// When updating with struct, GORM will only update non-zero fields. You might want to use map to update attributes or use Select to specify fields to update
// Default value도 updated 되게 끔
func (s Storage) UpdateAssetInfo(asset m.Asset) error {

	result := s.db.Select("*").Updates(asset)

	if result.Error != nil {
		return result.Error
	}

	s.lg.Info().Msgf("Updated asset with ID %d", asset.ID)
	return nil
}

func (s Storage) DeleteAssetInfo(id uint) error {

	result := s.db.Delete(&m.Asset{}, id)

	if result.Error != nil {
		return result.Error
	}

	s.lg.Info().Msgf("Soft deleted asset with ID %d", id)
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
		result := s.db.Where("created_at = ?", date).Last(&market) // Preload("Asset")
		if result.Error != nil {
			return nil, result.Error
		}
	}

	s.lg.Info().Msgf("Retrieved market status for date %s", date)
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

	s.lg.Info().Msgf("Retrieved market indicator for date %s", date)
	return &dailyIdx, &cliIdx, nil
}

func (s Storage) RetrieveMarketIndicatorWeekDesc() ([]m.DailyIndex, error) {

	var indexes []m.DailyIndex

	// endDate := time.Now().Format("2006-01-02")
	// startDate := time.Now().AddDate(0, 0, -7).Format("2006-01-02")
	// err := s.db.Where("created_at BETWEEN ? AND ?", startDate, endDate).
	// 	Find(&indexes).
	// 	Error

	err := s.db.Order("created_at DESC").
		Limit(7).
		Find(&indexes).
		Error

	s.lg.Info().Msgf("Retrieved %d market indicators for the last week", len(indexes))
	return indexes, err

}

func (s Storage) SaveDailyMarketIndicator(fearGreedIndex uint, nasdaq float64, sp500 float64) error {

	result := s.db.Create(&m.DailyIndex{
		CreatedAt:      time.Now(),
		FearGreedIndex: fearGreedIndex,
		NasDaq:         nasdaq,
		Sp500:          sp500,
	})
	if result.Error != nil {
		return result.Error
	}

	s.lg.Info().Msgf("Saved daily market indicator")
	return nil
}

func (s Storage) SaveMarketStatus(status uint) error {

	result := s.db.Create(&m.Market{
		CreatedAt: time.Now(),
		Status:    status,
	})
	if result.Error != nil {
		return result.Error
	}

	s.lg.Info().Msgf("Saved market status")
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

	s.lg.Info().Msgf("Retrieved %d invest history records", len(investHist))
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

	s.lg.Info().Msgf("Saved invest record for fund ID %d and asset ID %d", fundId, assetId)
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

	s.lg.Info().Msgf("Retrieved invest summary for fund ID %d and asset ID %d", fundId, assetId)
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

	s.lg.Info().Msgf("Updated invest summary for fund ID %d and asset ID %d", fundId, assetId)
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
	s.lg.Info().Msgf("Updated invest summary sum for fund ID %d and asset ID %d", fundId, assetId)
	return nil
}

func (s Storage) RetreiveLatestEma(assetId uint) (*m.EmaHist, error) {

	var ema m.EmaHist
	// result := s.db.Where("asset_id", assetId).Order("date desc").First(&ema)
	result := s.db.Where("asset_id", assetId).Last(&ema)
	if result.Error != nil {
		return nil, result.Error
	}

	s.lg.Info().Msgf("Retrieved latest EMA for asset ID %d", assetId)
	return &ema, nil
}

func (s Storage) SaveEmaHist(newEma *m.EmaHist) error {

	newEma.Date = time.Now()

	result := s.db.Create(newEma)
	if result.Error != nil {
		return result.Error
	}

	s.lg.Info().Msgf("Saved EMA history for asset ID %d", newEma.AssetID)
	return nil
}

func (s Storage) User(userName string) (*m.User, error) {

	var user m.User
	result := s.db.Where("username", userName).Last(&user)
	if result.Error != nil {
		return nil, result.Error
	}

	s.lg.Info().Msgf("Retrieved user with username %s", userName)
	return &user, nil
}

func (s Storage) RetreiveEventIsActive(eventId uint) bool {
	var event m.Event
	result := s.db.Where("id", eventId).First(&event)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			s.db.Create(&m.Event{ID: eventId, IsActive: true})
			s.lg.Info().Msgf("Created new event with ID %d and set as active", eventId)
			return true
		} else {
			s.lg.Info().Msgf("Failed to retrieve event with ID %d", eventId)
			return false
		}
	}

	s.lg.Info().Msgf("Retrieved event with ID %d, active status: %t", eventId, event.IsActive)
	return event.IsActive
}

func (s Storage) UpdateEventIsActive(eventId uint, isActive bool) error {

	result := s.db.Select("is_active").Updates(m.Event{ID: eventId, IsActive: isActive})
	if result.Error != nil {
		return result.Error
	}

	s.lg.Info().Msgf("Updated event with ID %d to active status: %t", eventId, isActive)
	return nil
}

func (s Storage) RetrieveLatestHighYieldSpread() (*m.HighYieldSpread, error) {
	var hy m.HighYieldSpread

	result := s.db.Last(&hy)
	if result.Error != nil {
		return nil, result.Error
	}

	return &hy, nil
}

func (s Storage) SaveHighYieldSpread(hy *m.HighYieldSpread) error {

	result := s.db.Create(hy)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (s Storage) RetrieveHighYieldSpreadWeekDesc() ([]m.HighYieldSpread, error) {
	var hy []m.HighYieldSpread

	err := s.db.Order("created_at DESC").
		Limit(7).
		Find(&hy).
		Error

	return hy, err
}

func (s Storage) RetrieveLatestSP500Entry() (*m.SP500Company, error) {
	var sp500 m.SP500Company

	result := s.db.Order("date_added DESC").First(&sp500)
	if result.Error != nil {
		return nil, result.Error
	}

	return &sp500, nil
}

// func (s Storage) RetrieveInitAmountofAsset(fundId, assetId uint) (float64, error) {

// 	var invests []m.Invest

// 	result := s.db.Model(&m.Invest{}).
// 		Where("fund_id", fundId).
// 		Where("asset_id", assetId).
// 		Find(&invests)

// 	if result.Error != nil {
// 		return 0, result.Error
// 	}

// 	rtn := 0.0

// 	for _, i := range invests {
// 		rtn += i.Count * i.Price
// 	}

// 	return rtn, nil
// }
