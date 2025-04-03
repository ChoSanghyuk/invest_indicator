package scrape

import (
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGoldApi(t *testing.T) {

	url := "https://www.goldapi.io/api/XAU/USD"
	key := os.Getenv("gold_key")
	head := map[string]string{
		"x-access-token": key,
	}

	var rtn map[string]interface{}
	err := sendRequest(url, http.MethodGet, head, nil, &rtn)
	if err != nil {
		t.Error(err)
	}
	assert.NotEmpty(t, rtn)
	t.Log(rtn)

	p := rtn["price_gram_24k"].(float64)
	t.Log(p)
}

func TestBitcoinApi(t *testing.T) {

	s := NewScraper(transmitterMock{})

	pp, cp, err := s.upbitApi("KRW-BTC")
	if err != nil {
		t.Error(err)
	}

	t.Logf("현재가 : %f\n시가: %f", pp, cp)
}

func TestAlpaca(t *testing.T) {
	pp, err := AlpacaCrypto("BTC/USD")
	if err != nil {
		t.Error(err)
	}
	t.Log(pp)
}

func TestGoldCrwal(t *testing.T) {

	s := Scraper{}

	// gold
	url := ""
	cssPath := ""

	rtn, err := s.crawl(url, cssPath)
	if err != nil {
		t.Error(err)
	}

	assert.NotEmpty(t, rtn)

	t.Log(rtn)
}

func TestBitcoinCrwal(t *testing.T) {

	s := Scraper{}

	t.Run("Crwal", func(t *testing.T) {
		url := ""
		cssPath := ""

		rtn, err := s.crawl(url, cssPath)
		if err != nil {
			t.Error(err)
		}

		assert.NotEmpty(t, rtn)

		t.Log(rtn)
	})

}

func TestEstateCrwal(t *testing.T) {

	s := Scraper{}

	t.Run("Crwal", func(t *testing.T) {
		url := ""
		cssPath := ""

		rtn, err := s.crawl(url, cssPath)
		if err != nil {
			t.Error(err)
		}

		assert.NotEmpty(t, rtn)
		assert.Equal(t, "예정지구 지정", rtn)
		t.Log(rtn)
	})
}

func TestExchangeRate(t *testing.T) {

	s := NewScraper(transmitterMock{})
	exrate := s.ExchageRate()
	t.Log(exrate)
}

func TestFearGreedIndex(t *testing.T) {

	s := NewScraper(transmitterMock{})

	rtn, err := s.FearGreedIndex()
	if err != nil {
		t.Error(err)
	}
	t.Logf("\n%+v", rtn)
}

func TestCliIndex(t *testing.T) {

	crwalByChromedp()
}
