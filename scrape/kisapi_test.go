package scrape

import (
	"os"
	"testing"
)

// todo. test 작성
type transmitterMock struct {
}

func (t transmitterMock) ApiBaseUrl(target string) string {
	return ""
}
func (t transmitterMock) ApiHeader(target string) map[string]string {
	return nil
}
func (t transmitterMock) CrawlUrlCasspath(target string) (url string, cssPath string) {
	return "", ""
}

func TestKis(t *testing.T) {

	appkey := os.Getenv("appkey")
	appsecret := os.Getenv("appsecret")
	token := os.Getenv("token")
	s := NewScraper(
		transmitterMock{},
		WithKIS(&KisConfig{
			AppKey:    appkey,
			AppSecret: appsecret,
		}),
		WithToken(token),
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

	t.Run("Nasdaq Index", func(t *testing.T) {
		pp, err := s.kisIndex(Nasdaq)
		if err != nil {
			t.Error(err)
		}
		t.Log(pp)
	})

	t.Run("S&P 500 Index", func(t *testing.T) {
		pp, err := s.kisIndex(Sp500)
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
