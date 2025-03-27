package model

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Fund struct {
	ID   uint
	Name string
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
	Date    datatypes.Date
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
	CreatedAt datatypes.Date
}

type DailyIndex struct {
	CreatedAt      datatypes.Date `gorm:"primaryKey"`
	FearGreedIndex uint
	NasDaq         float64
	Sp500          float64
}

type CliIndex struct {
	CreatedAt datatypes.Date `gorm:"primaryKey"`
	Index     float64
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
}

type Event struct {
	ID       uint
	IsActive bool
}
