package db

import (
	"fmt"
	m "investindicator/internal/model"
	"testing"
	"time"

	"gorm.io/datatypes"
)

func TestMigration(t *testing.T) {
	// db.AutoMigrate(&m.EmaHist{})
	stg.db.AutoMigrate(&m.Fund{}, &m.Asset{}, &m.Invest{}, &m.InvestSummary{}, &m.Market{}, &m.DailyIndex{}, &m.CliIndex{}, &m.EmaHist{}, &m.User{}, &m.Event{})
}

func TestCreate(t *testing.T) {
	fund := m.Fund{
		Name: "개인",
	}

	result := stg.db.Create(&fund)

	if result.Error != nil {
		t.Fatal(result.Error)
	}
	t.Log("ID", fund.ID)
	t.Log("Rows Affected", result.RowsAffected)
}

func TestRetrieve(t *testing.T) {
	var asset m.Asset

	result := stg.db.Model(&m.Asset{}).Where("id", 99).Find(&asset)
	fmt.Println(result.RowsAffected)
	if result.Error != nil || result.RowsAffected == 0 {
		return
	}
}

/*
결국은 time.Time 객체인 것이 중요한게 아닌, string형 변환했을 때 DB 타입과 일치하는지가 중요함
time.Time{}.Local() => '0000-00-00 00:00:00' 라서 Date 타입 및 Timestamp 실패
time.Now() => '2024-08-16 08:47:20.346' Date 타입 및 Timestamp 성공
*/
func TestTime(t *testing.T) {
	stg.db.AutoMigrate(&m.Sample{})

	// date, _ := time.Parse("2006-01-02", "2021-11-22")

	d := m.Sample{
		Date: datatypes.Date(time.Now()),
		Time: time.Now(),
	}

	stg.db.Debug().Create(&d)
}

func TestSelectFirst(t *testing.T) {
	var dailyIdx m.DailyIndex

	result := stg.db.Where("created_at = ?", "2024-09-21").Select(&dailyIdx)
	if result.Error != nil {
		t.Error(result.Error)
	}

	fmt.Printf("%+v", dailyIdx)
}

func TestRetrieveInvestSummary(t *testing.T) {

	fundId := 1
	assetId := 12

	var investSummary m.InvestSummary
	result := stg.db.Model(&m.InvestSummary{}).
		Where("fund_id = ?", fundId).
		Where("asset_id = ?", assetId).
		Find(&investSummary)

	if result.RowsAffected == 0 {
		t.Error("RowsAffected : 0")
	}
	t.Log(investSummary)
}
