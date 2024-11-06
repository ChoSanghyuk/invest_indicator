package db

import (
	"database/sql"
	"fmt"
	m "invest/model"
	"log"
	"testing"
	"time"

	"gorm.io/datatypes"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

func init() {

	dsn := "root:root@tcp(127.0.0.1:3306)/investdb?charset=utf8mb4&parseTime=True&loc=Local"
	sqlDB, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}

	db, err = gorm.Open(mysql.New(mysql.Config{
		Conn: sqlDB,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal(err)
	}
}

func TestMigration(t *testing.T) {
	// db.AutoMigrate(&m.EmaHist{})
	db.AutoMigrate(&m.Fund{}, &m.Asset{}, &m.Invest{}, &m.InvestSummary{}, &m.Market{}, &m.DailyIndex{}, &m.CliIndex{}, &m.EmaHist{})
}

func TestCreate(t *testing.T) {
	fund := m.Fund{
		Name: "개인",
	}

	result := db.Create(&fund)

	if result.Error != nil {
		t.Fatal(result.Error)
	}
	t.Log("ID", fund.ID)
	t.Log("Rows Affected", result.RowsAffected)
}

func TestCreateAsset(t *testing.T) {

	_, err := stg.SaveAssetInfo("bitcoin", m.DomesticCoin, "KRW-BTC", "WON", 98000000, 68000000, 88000000, 70000000)
	if err != nil {
		t.Error(err)
	}

	_, err = stg.SaveAssetInfo("gold", m.Gold, "M04020000", "WON", 111360, 80100, 0, 103630)
	if err != nil {
		t.Error(err)
	}
}

func TestRetrieve(t *testing.T) {
	var asset m.Asset

	result := db.Model(&m.Asset{}).Where("id", 99).Find(&asset)
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
	db.AutoMigrate(&m.Sample{})

	// date, _ := time.Parse("2006-01-02", "2021-11-22")

	d := m.Sample{
		Date: datatypes.Date(time.Now()),
		Time: time.Now(),
	}

	db.Debug().Create(&d)
}

func TestSelectFirst(t *testing.T) {
	var dailyIdx m.DailyIndex

	result := db.Where("created_at = ?", "2024-09-21").Select(&dailyIdx)
	if result.Error != nil {
		t.Error(result.Error)
	}

	fmt.Printf("%+v", dailyIdx)
}
