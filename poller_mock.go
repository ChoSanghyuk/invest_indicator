package investind

import (
	md "investindicator/internal/model"
)

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

func (m RtPollerMock) GoldPriceDollar() (float64, error) {
	return 0, nil
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
func (m DailyPollerMock) Sp500() (float64, error) {
	return 0, nil
}
func (m DailyPollerMock) CliIdx() (float64, error) {
	return 0, nil
}

func (m DailyPollerMock) ClosingPrice(category md.Category, code string) (float64, error) {
	return 0, nil
}
func (m DailyPollerMock) HighYieldSpread() (date string, spread float64, err error) {
	return "", 0, nil
}
