package scrape

import (
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGoldApi(t *testing.T) {

	_ = "https://www.goldapi.io/api/XAU/USD"
	key := os.Getenv("gold_key")
	head := map[string]string{
		"x-access-token": key,
	}

	var rtn map[string]interface{}
	err := sendRequest(goldPriceDollarUrl, http.MethodGet, head, nil, &rtn)
	if err != nil {
		t.Error(err)
	}
	assert.NotEmpty(t, rtn)
	t.Log(rtn)

	p := rtn["price_gram_24k"].(float64)
	t.Log(p)
}

func TestBitcoinApi(t *testing.T) {

	s, err := NewScraper(transmitterMock{})
	if err != nil {
		t.Error(err)
	}

	pp, cp, err := s.upbitApi("KRW-BTC")
	if err != nil {
		t.Error(err)
	}

	t.Logf("현재가 : %f\n시가: %f", pp, cp)
}

func TestAlpaca(t *testing.T) {
	pp, err := alpacaCrypto("BTC")
	if err != nil {
		t.Error(err)
	}
	t.Log(pp)
}

func TestGoldCrwal(t *testing.T) {

	s := Scraper{}

	// gold
	url := "https://data-as.goldprice.org/dbXRates/USD"
	cssPath := "#goldchange > div > div > div > div > div.tick-value-wrap.d-flex.align-items-center.justify-content-center.flex-wrap > div.tick-value.price-value > span"

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

	s, _ := NewScraper(transmitterMock{})
	exrate := s.ExchageRate()
	t.Log(exrate)
}

func TestFearGreedIndex(t *testing.T) {

	s, _ := NewScraper(transmitterMock{})

	rtn, err := s.FearGreedIndex()
	if err != nil {
		t.Error(err)
	}
	t.Logf("\n%+v", rtn)
}

func TestCliIndex(t *testing.T) {

	crwalByChromedp()
}

func TestHighYieldSpread(t *testing.T) {

	s, _ := NewScraper(transmitterMock{})

	date, spread, err := s.HighYieldSpread()
	if err != nil {
		t.Error(err)
	}
	t.Log(date, spread)
}

func TestSP500List(t *testing.T) {

	s, _ := NewScraper(transmitterMock{})

	entry, err := s.RecentSP500Entries("2025-07-01")
	if err != nil {
		t.Error(err)
	}
	t.Log(entry)
}

func TestNewScraper(t *testing.T) {
	
	t.Run("NewScraper with transmitter only", func(t *testing.T) {
		s, err := NewScraper(transmitterMock{})
		
		assert.NoError(t, err)
		assert.NotNil(t, s)
		assert.NotNil(t, s.t)
		assert.NotNil(t, s.lg)
	})

	t.Run("NewScraper with valid KIS option", func(t *testing.T) {
		kisConfig := &KisConfig{
			AppKey:    "test_appkey",
			AppSecret: "test_appsecret", 
			Account:   "test_account",
		}
		
		s, err := NewScraper(transmitterMock{}, WithKIS(kisConfig))
		
		assert.NoError(t, err)
		assert.NotNil(t, s)
		assert.Equal(t, "test_appkey", s.kis.appKey)
		assert.Equal(t, "test_appsecret", s.kis.appSecret)
		assert.Equal(t, "test_account", s.kis.account)
	})

	t.Run("NewScraper with valid Token option", func(t *testing.T) {
		token := "test_token"
		
		s, err := NewScraper(transmitterMock{}, WithToken(token))
		
		assert.NoError(t, err)
		assert.NotNil(t, s)
		assert.Equal(t, token, s.kis.accessToken)
		assert.NotEmpty(t, s.kis.tokenExpired)
	})

	t.Run("NewScraper with multiple valid options", func(t *testing.T) {
		kisConfig := &KisConfig{
			AppKey:    "test_appkey",
			AppSecret: "test_appsecret",
			Account:   "test_account",
		}
		token := "test_token"
		
		s, err := NewScraper(transmitterMock{}, WithKIS(kisConfig), WithToken(token))
		
		assert.NoError(t, err)
		assert.NotNil(t, s)
		assert.Equal(t, "test_appkey", s.kis.appKey)
		assert.Equal(t, "test_appsecret", s.kis.appSecret)
		assert.Equal(t, "test_account", s.kis.account)
		assert.Equal(t, token, s.kis.accessToken)
		assert.NotEmpty(t, s.kis.tokenExpired)
	})

	t.Run("NewScraper with invalid KIS option - missing AppKey", func(t *testing.T) {
		kisConfig := &KisConfig{
			AppKey:    "", // empty
			AppSecret: "test_appsecret",
			Account:   "test_account",
		}
		
		s, err := NewScraper(transmitterMock{}, WithKIS(kisConfig))
		
		assert.Error(t, err)
		assert.Nil(t, s)
		assert.Contains(t, err.Error(), "kis appkey 미존재")
	})

	t.Run("NewScraper with invalid KIS option - missing AppSecret", func(t *testing.T) {
		kisConfig := &KisConfig{
			AppKey:    "test_appkey",
			AppSecret: "", // empty
			Account:   "test_account",
		}
		
		s, err := NewScraper(transmitterMock{}, WithKIS(kisConfig))
		
		assert.Error(t, err)
		assert.Nil(t, s)
		assert.Contains(t, err.Error(), "kis appsecret 미존재")
	})

	t.Run("NewScraper with invalid KIS option - missing Account", func(t *testing.T) {
		kisConfig := &KisConfig{
			AppKey:    "test_appkey",
			AppSecret: "test_appsecret",
			Account:   "", // empty
		}
		
		s, err := NewScraper(transmitterMock{}, WithKIS(kisConfig))
		
		assert.Error(t, err)
		assert.Nil(t, s)
		assert.Contains(t, err.Error(), "kis accoount 미존재")
	})
}
