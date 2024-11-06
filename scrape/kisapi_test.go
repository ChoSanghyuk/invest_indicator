package scrape

import (
	"invest/config"
	"testing"
)

func TestKis(t *testing.T) {

	conf, err := config.NewConfig()
	if err != nil {
		panic(err)
	}
	conf.InitKIS("")

	s := NewScraper(
		conf,
		WithKIS(conf.KisAppKey(), conf.KisAppSecret()),
	)

	t.Run("Token Generate", func(t *testing.T) {
		token, err := s.KisToken()
		if err != nil {
			t.Error(err)
		}
		t.Log(token)
	})

	t.Run("Stock current Price", func(t *testing.T) {
		stock, err := s.kisDomesticStockPrice("M04020000")
		if err != nil {
			t.Error(err)
		}
		t.Log(stock.pp, stock.op, stock.hp, stock.lp, stock.ap)
	})

	t.Run("Foreign stock", func(t *testing.T) {
		pp, cp, err := s.kisForeignPrice("NAS-MSFT")
		if err != nil {
			t.Error(err)
		}
		t.Log(pp, cp, err)
	})

	t.Run("Foreign Index", func(t *testing.T) {
		pp, err := s.kisNasdaqIndex()
		if err != nil {
			t.Error(err)
		}
		t.Log(pp)
	})

	t.Run("Domestic ETF", func(t *testing.T) {
		stock, err := s.kisDomesticEtfPrice("360750")
		if err != nil {
			t.Error(err)
		}
		t.Log(stock.pp, stock.op, stock.hp, stock.lp, stock.ap)
	})

	t.Run("Foreign ETF", func(t *testing.T) {
		pp, cp, err := s.kisForeignPrice("AMS-SPY")
		if err != nil {
			t.Error(err)
		}
		t.Log(pp, cp)
	})

}
