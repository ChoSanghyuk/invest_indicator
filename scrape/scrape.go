package scrape

import (
	"errors"
	"fmt"
	m "investindicator/internal/model"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
	"github.com/gofiber/fiber/v2/log"
	"github.com/rs/zerolog"
)

type Scraper struct {
	exchange struct {
		Rate float64
		Date time.Time
	}
	kis struct {
		appKey       string
		appSecret    string
		accessToken  string
		tokenExpired string
	}
	lg zerolog.Logger
	t  transmitter
}

type transmitter interface {
	Key(target string) string
}

func NewScraper(t transmitter, options ...func(*Scraper)) *Scraper {
	s := &Scraper{
		t:  t,
		lg: zerolog.New(os.Stdout).With().Str("Module", "Scraper").Timestamp().Logger(),
	}
	for _, opt := range options {
		opt(s)
	}
	return s
}

type KisConfig struct {
	AppKey    string
	AppSecret string
}

func WithKIS(conf *KisConfig) func(*Scraper) {

	return func(s *Scraper) {
		s.kis.appKey = conf.AppKey
		s.kis.appSecret = conf.AppSecret
	}
}

func WithToken(token string) func(*Scraper) {

	return func(s *Scraper) {
		s.kis.accessToken = token
		s.kis.tokenExpired = time.Now().Add(time.Duration(1) * time.Hour).Format("2006-01-02 15:04:05")
	}
}

/*
종목 이름만 보고 어디서 가져올 지 정할 수 있어야 함
종목별로 타입을 지정 => 어떤 base url을 사용할 지 결정

	어떤 base url일지는 PresentPrice 내부에서 case 세분화

종목별로 심볼 등 base url에 들어갈 인자를 정할 수 있어야함

종목 이름 - 타입/심볼을 어디에 저장해 둘 것인가 => DB
*/

func (s *Scraper) PresentPrice(category m.Category, code string) (pp float64, err error) {
	s.lg.Info().Msgf("Starting PresentPrice with category: %v, code: %s", category, code)
	switch category {
	case m.Won:
		return 1, nil
	case m.Dollar:
		return s.ExchageRate(), nil
	case m.DomesticStock, m.Gold:
		stock, err := s.kisDomesticStockPrice(code)
		return stock.pp, err
	case m.DomesticETF:
		stock, err := s.kisDomesticEtfPrice(code)
		return stock.pp, err
	case m.DomesticCoin:
		pp, _, err := s.bithumbApi(code)
		return pp, err
	case m.ForeignCoin:
		pp, err := alpacaCrypto(code)
		return pp, err
	case m.ForeignStock, m.ForeignETF:
		pp, _, err := s.kisForeignPrice(code)
		return pp, err
	}

	s.lg.Error().Err(err).Msg("Error in PresentPrice")
	return 0, errors.New("미분류된 종목")
}

func (s *Scraper) TopBottomPrice(category m.Category, code string) (hp float64, lp float64, err error) {
	s.lg.Info().Msgf("Starting TopBottomPrice with category: %v, code: %s", category, code)
	switch category {
	case m.DomesticStock:
		stock, err := s.kisDomesticStockPrice(code)
		return stock.hp, stock.lp, err
	case m.DomesticETF:
		stock, err := s.kisDomesticEtfPrice(code)
		return stock.hp, stock.lp, err
		// case model.DomesticCoin:
		// 	return 0, 0, nil
	}

	s.lg.Error().Err(err).Msg("Error in TopBottomPrice")
	return 0, 0, errors.New("최고/최저 호출 API 미존재")
}

func (s *Scraper) AvgPrice(category m.Category, code string) (float64, uint, error) {
	s.lg.Info().Msgf("Starting AvgPrice with category: %v, code: %s", category, code)
	switch category {
	case m.DomesticStock:
		stock, err := s.kisDomesticStockPrice(code)
		return stock.ap, 200, err
	case m.ForeignStock:
		ap, n, err := s.kisForeignAvg(code)
		return ap, uint(n), err
	}

	// s.lg.Error().Err(err).Msg("Error in AvgPrice")
	return 0, 0, errors.New("평균 가격 호출 API 미존재")
}

func (s *Scraper) ClosingPrice(category m.Category, code string) (cp float64, err error) {
	s.lg.Info().Msgf("Starting ClosingPrice with category: %v, code: %s", category, code)
	switch category {
	case m.Won:
		return 1, nil
	case m.Dollar:
		return s.ExchageRate(), nil
	case m.DomesticStock, m.Gold:
		stock, err := s.kisDomesticStockPrice(code)
		return stock.op, err
	case m.DomesticCoin:
		_, cp, err = s.bithumbApi(code)
		return cp, err
	case m.DomesticETF:
		stock, err := s.kisDomesticEtfPrice(code)
		return stock.op, err
	case m.ForeignStock, m.ForeignETF:
		_, cp, err := s.kisForeignPrice(code)
		return cp, err
	}

	s.lg.Error().Err(err).Msg("Error in ClosingPrice")
	return 0, errors.New("미분류된 종목")
}

const realEstateUrl = "https://www.ep.go.kr/www/contents.do?key=3763"
const realEstateCss = "#contents > table:nth-child(8) > tbody > tr:nth-child(2) > td:nth-child(6)"

func (s *Scraper) RealEstateStatus() (string, error) {
	s.lg.Info().Msg("Starting RealEstateStatus")
	return s.crawl(realEstateUrl, realEstateCss)
}

const exRateUrl = "https://search.naver.com/search.naver?where=nexearch&sm=top_hty&fbm=0&ie=utf8&query=%ED%99%98%EC%9C%A8"
const exRateCssPath = "#main_pack > section.sc_new.cs_nexchangerate > div:nth-child(1) > div.exchange_bx._exchange_rate_calculator > div > div.inner > div:nth-child(3) > div.num > div > span"

func (s *Scraper) ExchageRate() float64 {
	s.lg.Info().Msg("Starting ExchageRate")
	if s.exchange.Rate != 0 && !s.exchange.Date.Before(time.Now().Add(-3*time.Hour)) {
		return s.exchange.Rate
	}

	rtn, err := s.crawl(exRateUrl, exRateCssPath)
	if err != nil {
		log.Error(err)
	}

	re := regexp.MustCompile(`[^\d.]`)
	exrate, err := strconv.ParseFloat(re.ReplaceAllString(rtn, ""), 64)
	if err != nil {
		return 0
	}

	s.exchange.Rate = exrate
	s.exchange.Date = time.Now()

	return exrate
}

const (
	fearGreedUrl = "https://fear-and-greed-index.p.rapidapi.com/v1/fgi"
)

func (s *Scraper) FearGreedIndex() (uint, error) {
	s.lg.Info().Msg("Starting FearGreedIndex")

	key := s.t.Key("rapidapi")
	header := map[string]string{
		"x-rapidapi-host": "fear-and-greed-index.p.rapidapi.com",
		"x-rapidapi-key":  key,
	}

	type fearGreed struct {
		Fgi struct {
			Now struct {
				Value uint   `json:"value"`
				Text  string `json:"valueText"`
			} `json:"now"`
		} `json:"fgi"`
	}
	var rtn fearGreed

	err := sendRequest(fearGreedUrl, http.MethodGet, header, nil, &rtn)
	if err != nil {
		return 0, nil
	}

	return rtn.Fgi.Now.Value, nil
}

func (s *Scraper) Nasdaq() (float64, error) {
	s.lg.Info().Msg("Starting Nasdaq")
	return s.kisIndex(Nasdaq)
}

func (s *Scraper) Sp500() (float64, error) {
	s.lg.Info().Msg("Starting Sp500")
	return s.kisIndex(Sp500)
}

// todo. 현재로는 크롤링/API 못 찾음
func (s *Scraper) CliIdx() (float64, error) {
	s.lg.Info().Msg("Starting CliIdx")
	// need Chromedp
	return 0, nil
}

const (
	goldPriceDollarUrl = "https://www.goldapi.io/api/XAU/USD"
)

func (s *Scraper) GoldPriceDollar() (float64, error) {
	s.lg.Info().Msg("Starting GoldPriceDollar")

	key := s.t.Key("goldapi")

	head := map[string]string{
		"x-access-token": key,
	}

	var rtn map[string]interface{}
	err := sendRequest(goldPriceDollarUrl, http.MethodGet, head, nil, &rtn)
	if err != nil {
		return 0, err
	}

	if rtn["error"] != nil {
		return 0, fmt.Errorf("%s", rtn["error"])
	}

	p := rtn["price_gram_24k"].(float64)

	return p, nil
}

// todo. refactor scraper 변경 필요
func (s *Scraper) Buy(category m.Category, code string) error {
	s.lg.Info().Msgf("Starting Buy with category: %v, code: %s", category, code)

	switch category {
	case m.ForeignStock, m.ForeignETF:
		err := s.kisForeignBuy(code, 0)
		return err
	}

	return nil
}

func alpacaCrypto(symbol string) (float64, error) {

	client := marketdata.NewClient(marketdata.ClientOpts{})
	request := marketdata.GetLatestCryptoBarRequest{}

	bar, err := client.GetLatestCryptoBar(symbol+"/USD", request)
	if err != nil {
		return 0, err
	}

	return bar.Close, nil
}
