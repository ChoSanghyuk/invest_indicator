package investind

import (
	m "investindicator/internal/model"
)

type rtPoller interface {
	PresentPrice(category m.Category, code string) (float64, error)
	RealEstateStatus() (string, error)
	GoldPriceDollar() (float64, error)
}

type dailyPoller interface {
	ExchageRate() float64
	ClosingPrice(category m.Category, code string) (float64, error)
	FearGreedIndex() (uint, error)
	Nasdaq() (float64, error)
	Sp500() (float64, error)
	CliIdx() (float64, error)
	HighYieldSpread() (date string, spread float64, err error)
}

type Poller interface {
	rtPoller
	dailyPoller
}
