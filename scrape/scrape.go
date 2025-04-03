package scrape

import (
	"encoding/base64"
	"errors"
	"fmt"
	m "invest/model"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
	"github.com/gofiber/fiber/v2/log"
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
	t transmitter
}

type transmitter interface {
	ApiBaseUrl(target string) string
	ApiHeader(target string) map[string]string
	CrawlUrlCasspath(target string) (url string, cssPath string)
}

func NewScraper(t transmitter, options ...func(*Scraper)) *Scraper {
	s := &Scraper{
		t: t,
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

	switch category {
	case m.Won:
		return 1, nil
	case m.Dollar:
		return 1, nil
	case m.DomesticStock, m.Gold:
		stock, err := s.kisDomesticStockPrice(code)
		return stock.pp, err
	case m.DomesticETF:
		stock, err := s.kisDomesticEtfPrice(code)
		return stock.pp, err
	case m.DomesticCoin:
		pp, _, err := s.upbitApi(code)
		return pp, err
	case m.ForeignStock, m.ForeignETF:
		pp, _, err := s.kisForeignPrice(code)
		return pp, err
	}

	return 0, errors.New("미분류된 종목")
}

func (s *Scraper) TopBottomPrice(category m.Category, code string) (hp float64, lp float64, err error) {
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

	return 0, 0, errors.New("최고/최저 호출 API 미존재")
}

func (s *Scraper) AvgPrice(category m.Category, code string) (float64, uint, error) {
	switch category {
	case m.DomesticStock:
		stock, err := s.kisDomesticStockPrice(code)
		return stock.ap, 200, err
	case m.DomesticETF:
		ap, n, err := s.kisForeignAvg(code)
		return ap, uint(n), err
	}

	return 0, 0, errors.New("평균 가격 호출 API 미존재")
}

func (s *Scraper) ClosingPrice(category m.Category, code string) (cp float64, err error) {

	switch category {
	case m.Won:
		return 1, nil
	case m.Dollar:
		r := s.ExchageRate()
		return r, nil
	case m.DomesticStock, m.Gold:
		stock, err := s.kisDomesticStockPrice(code)
		return stock.op, err
	case m.DomesticCoin:
		_, cp, err = s.upbitApi(code)
		return cp, err
	case m.DomesticETF:
		stock, err := s.kisDomesticEtfPrice(code)
		return stock.op, err
	case m.ForeignStock, m.ForeignETF:
		_, cp, err := s.kisForeignPrice(code)
		return cp, err
	}

	return 0, errors.New("미분류된 종목")
}

func (s *Scraper) RealEstateStatus() (string, error) {
	url, cssPath := s.t.CrawlUrlCasspath("estate")
	return s.crawl(url, cssPath)
}

func (s *Scraper) ExchageRate() float64 {

	if s.exchange.Rate != 0 && s.exchange.Date.Format("20060102") == time.Now().Format("20060102") {
		return s.exchange.Rate
	}

	url, cssPath := s.t.CrawlUrlCasspath("exchangeRate")

	rtn, err := s.crawl(url, cssPath)
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

func (s *Scraper) FearGreedIndex() (uint, error) {

	url := s.t.ApiBaseUrl("fearGreed")
	header := s.t.ApiHeader("fearGreed")

	type fearGreed struct {
		Fgi struct {
			Now struct {
				Value uint   `json:"value"`
				Text  string `json:"valueText"`
			} `json:"now"`
		} `json:"fgi"`
	}
	var rtn fearGreed

	err := sendRequest(url, http.MethodGet, header, nil, &rtn)
	if err != nil {
		return 0, nil
	}

	return rtn.Fgi.Now.Value, nil
}

func (s *Scraper) Nasdaq() (float64, error) {

	return s.kisIndex(Nasdaq)
}

func (s *Scraper) Sp500() (float64, error) {

	return s.kisIndex(Sp500)
}

// todo. 현재로는 크롤링/API 못 찾음
func (s *Scraper) CliIdx() (float64, error) {
	// need Chromedp
	return 0, nil
}

func (s *Scraper) GoldPriceDollar() (float64, error) {
	url := "https://www.goldapi.io/api/XAU/KRW" // USD

	b64Key := s.t.ApiHeader("gold")["api-key"] // todo. confing Key transmitter 개설
	key, _ := base64.StdEncoding.DecodeString(b64Key)
	head := map[string]string{
		"x-access-token": string(key),
	}

	var rtn map[string]interface{}
	err := sendRequest(url, http.MethodGet, head, nil, &rtn)
	if err != nil {
		return 0, err
	}

	p := rtn["price_gram_24k"].(float64)

	return p, nil
}

// todo. refactor scraper 변경 필요
func (s *Scraper) Buy(category m.Category, code string) error {

	switch category {
	case m.ForeignStock, m.ForeignETF:
		err := s.kisForeignBuy(code, 0)
		return err
	}

	return nil
}

// depre
func AlpacaCrypto(target string) (string, error) {

	client := marketdata.NewClient(marketdata.ClientOpts{})
	request := marketdata.GetCryptoBarsRequest{
		TimeFrame: marketdata.OneMin,
		Start:     time.Now().Add(time.Duration(-10) * time.Minute), // time.Date(2022, 9, 1, 0, 0, 0, 0, time.UTC),
		End:       time.Now(),
	}

	bars, err := client.GetCryptoBars(target, request)
	if err != nil {
		return "", err
	}

	if len(bars) == 0 {
		return "", errors.New("빈 결과값 반환")
	}

	bar := bars[len(bars)-1]
	return fmt.Sprintf("%f", bar.Close), nil
}
