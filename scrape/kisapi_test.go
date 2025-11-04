package scrape

import (
	"os"
	"testing"

	"github.com/rs/zerolog"
)

// todo. test 작성
type transmitterMock struct {
}

func (t transmitterMock) Key(target string) string {
	return ""
}

func TestKis(t *testing.T) {

	appkey := os.Getenv("appkey")
	appsecret := os.Getenv("appsecret")
	account := os.Getenv("account")
	token := os.Getenv("token")

	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	s, err := NewScraper(
		transmitterMock{},
		WithKIS(&KisConfig{
			AppKey:    appkey,
			AppSecret: appsecret,
			Account:   account,
		}),
		WithKisToken(token),
	)
	if err != nil {
		t.Fatal(err)
	}

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
		pp, cp, err := s.kisForeignPrice("AMS-SPLG")
		if err != nil {
			t.Error(err)
		}
		t.Log(pp, cp)
	})

	t.Run("Foreign Period Price", func(t *testing.T) {
		ap, n, err := s.kisForeignAvg("NAS-TSLA")
		if err != nil {
			t.Error(err)
		}
		t.Log(ap)
		t.Log(n)
	})

	t.Run("Domestic stock buy", func(t *testing.T) {
		err := s.kisDomesticBuy("024950", 1) // 삼천리 자전거
		if err != nil {
			t.Error(err)
		}
	})

}
