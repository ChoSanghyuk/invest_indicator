package model

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

/*
memo. time.Time은 datetime(3) 타입으로 저장됨. datatypes.Date은 date 타입으로 저장됨.
첫 생성이 datatypes.Date 였어도 그 이후에 필드의 타입을 time.Time으로 변경해서 사용해도 지장 X. 첫 필드의 타입처럼 date 타입으로 계속 값이 들어가고 나옴.
*/
type Fund struct {
	ID       uint
	Name     string
	IsExcept bool `gorm:"column:is_except;default:false"` // column mapping
}

type Asset struct {
	ID        uint
	Name      string
	Category  Category
	Code      string
	Currency  string
	Top       float64
	Bottom    float64
	SellPrice float64
	BuyPrice  float64
}

type EmaHist struct {
	ID      uint
	AssetID uint
	Asset   Asset `gorm:"foreignKey:AssetID;constraint:OnDelete:CASCADE"`
	Date    time.Time
	Ema     float64
	NDays   uint
}

type Invest struct {
	ID      uint
	FundID  uint
	Fund    Fund
	AssetID uint
	Asset   Asset
	Price   float64
	Count   float64
	gorm.Model
}

type InvestSummary struct {
	ID      uint
	FundID  uint
	Fund    Fund
	AssetID uint
	Asset   Asset
	Count   float64
	Sum     float64
}

type Market struct {
	ID        uint
	Status    uint
	CreatedAt time.Time
}

type DailyIndex struct {
	CreatedAt      time.Time `gorm:"primaryKey"`
	FearGreedIndex uint
	NasDaq         float64
	Sp500          float64
}

type CliIndex struct {
	CreatedAt time.Time `gorm:"primaryKey"`
	Index     float64
}

type HighYieldSpread struct {
	CreatedAt time.Time `gorm:"primaryKey"`
	Spread    float64
}

type Sample struct {
	ID   uint `gorm:"primaryKey"`
	Date datatypes.Date
	Time time.Time
	gorm.Model
}

type User struct {
	ID       int
	Username string
	Email    string
	Password string
	IsAdmin  bool
}

type Event struct {
	ID       uint
	IsActive bool
}

type SP500Company struct {
	ID                    uint      `json:"id" gorm:"primaryKey"`
	Symbol                string    `json:"symbol" gorm:"column:symbol"`
	Security              string    `json:"security" gorm:"column:security"`
	GICS_Sector           string    `json:"gics_sector" gorm:"column:gics_sector"`
	GICS_Sub_Industry     string    `json:"gics_sub_industry" gorm:"column:gics_sub_industry"`
	Headquarters_Location string    `json:"headquarters_location" gorm:"column:headquarters_location"`
	Date_added            time.Time `json:"date_added" gorm:"column:date_added;type:date"`
	CIK                   string    `json:"cik" gorm:"column:cik"`
	Founded               string    `json:"founded" gorm:"column:founded"`
}
