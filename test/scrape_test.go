package scrape

import (
	"invest/config"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGoldApi(t *testing.T) {

	info, _ := config.NewConfig()

	url := info.Api["gold"].Url
	head := info.Api["gold"].Header

	var rtn map[string]interface{}
	err := sendRequest(url, http.MethodGet, head, nil, rtn)
	if err != nil {
		t.Error(err)
	}

	assert.NotEmpty(t, rtn)

	t.Log(rtn)
}

func TestBitcoinApi(t *testing.T) {

	c, _ := config.NewConfig()
	c.InitKIS("")

	s := NewScraper(c)

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
	conf, _ := config.NewConfig()
	conf.InitKIS("")

	url := conf.Crawl["gold"].Url
	cssPath := conf.Crawl["gold"].CssPath

	rtn, err := s.crawl(url, cssPath)
	if err != nil {
		t.Error(err)
	}

	assert.NotEmpty(t, rtn)

	t.Log(rtn)
}

func TestBitcoinCrwal(t *testing.T) {

	s := Scraper{}
	conf, _ := config.NewConfig()
	conf.InitKIS("")

	t.Run("Crwal", func(t *testing.T) {
		url := conf.Crawl["bitcoin"].Url
		cssPath := conf.Crawl["bitcoin"].CssPath

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
	conf, _ := config.NewConfig()
	conf.InitKIS("")

	t.Run("Crwal", func(t *testing.T) {
		url := conf.Crawl["estate"].Url
		cssPath := conf.Crawl["estate"].CssPath

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

	conf, _ := config.NewConfig()
	conf.InitKIS("")

	s := NewScraper(conf)
	exrate := s.ExchageRate()
	t.Log(exrate)
}

func TestFearGreedIndex(t *testing.T) {

	conf, _ := config.NewConfig()
	conf.InitKIS("")

	s := NewScraper(conf)

	rtn, err := s.FearGreedIndex()
	if err != nil {
		t.Error(err)
	}
	t.Logf("\n%+v", rtn)
}

func TestCliIndex(t *testing.T) {

	// s := Scraper{}

	// url := "https://www.oecd.org/en/data/indicators/composite-leading-indicator-cli.html?oecdcontrol-b2a0dbca4d-var3=2008-01&oecdcontrol-b2a0dbca4d-var4=2024-12"
	// cssPath := "#highcharts-p773b0p-798 > svg > g.highcharts-series-group > g.highcharts-series.highcharts-series-0.highcharts-spline-series > path.highcharts-graph"

	// rtn, err := s.crawl(url, cssPath)
	// if err != nil {
	// 	t.Error(err)
	// }

	// assert.NotEmpty(t, rtn)

	// t.Log(rtn)
	crwalByChromedp()
}
