package scrape

import (
	"bytes"
	"errors"
	"fmt"
	m "investindicator/internal/model"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
	"github.com/gofiber/fiber/v2/log"
	"github.com/rs/zerolog"
)

type Scraper struct {
	exchange struct {
		Rate float64
		Date time.Time
	}
	kis   *Kis
	upbit struct {
		token string
	}
	lg zerolog.Logger
	t  transmitter // todo. 이거 제거 하자. 다 그냥 필드로 들고 있는 것으로.
}

type transmitter interface {
	Key(target string) string
}

type Option func(*Scraper) error

// Functional Option Pattern
func NewScraper(t transmitter, options ...Option) (*Scraper, error) {
	s := &Scraper{
		t:  t,
		lg: zerolog.New(os.Stdout).With().Str("Module", "Scraper").Timestamp().Logger(),
	}
	for _, opt := range options {
		if err := opt(s); err != nil {
			return nil, fmt.Errorf("failed to create Scraper %w", err)
		}
	}
	return s, nil
}

type KisConfig struct {
	AppKey    string
	AppSecret string
	Account   string
	HtsId     string
}

func WithKIS(conf *KisConfig) Option {

	return func(s *Scraper) error {
		if conf.AppKey == "" {
			return errors.New("kis appkey 미존재")
		}
		if conf.AppSecret == "" {
			return errors.New("kis appsecret 미존재")
		}
		if conf.Account == "" {
			return errors.New("kis accoount 미존재")
		}
		if conf.HtsId == "" {
			return errors.New("kis HTS ID 미존재")
		}

		s.kis = NewKis(conf.AppKey, conf.AppSecret, conf.Account, conf.HtsId)

		return nil
	}
}

func WithKisToken(token string) Option {

	return func(s *Scraper) error {
		if s.kis != nil {
			s.kis.SetAccessToken(token)
		}
		return nil
	}
}

// acount option 설정 추가

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
		stock, err := s.kis.DomesticStockPrice(code)
		return stock.pp, err
	case m.DomesticETF, m.DomesticGoldETF:
		stock, err := s.kis.DomesticEtfPrice(code)
		return stock.pp, err
	case m.DomesticCoin:
		pp, _, err := s.bithumbApi(code)
		return pp, err
	case m.ForeignCoin:
		pp, err := alpacaCrypto(code)
		return pp, err
	case m.ForeignStock, m.ForeignETF:
		pp, _, err := s.kis.ForeignPrice(code)
		return pp, err
	}

	s.lg.Error().Err(err).Msg("Error in PresentPrice")
	return 0, errors.New("미분류된 종목")
}

func (s *Scraper) TopBottomPrice(category m.Category, code string) (hp float64, lp float64, err error) {
	s.lg.Info().Msgf("Starting TopBottomPrice with category: %v, code: %s", category, code)
	switch category {
	case m.DomesticStock:
		stock, err := s.kis.DomesticStockPrice(code)
		return stock.hp, stock.lp, err
	case m.DomesticETF, m.DomesticGoldETF:
		stock, err := s.kis.DomesticEtfPrice(code)
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
		stock, err := s.kis.DomesticStockPrice(code)
		return stock.ap, 200, err
	case m.ForeignStock:
		ap, n, err := s.kis.ForeignAvg(code)
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
		stock, err := s.kis.DomesticStockPrice(code)
		return stock.op, err
	case m.DomesticCoin:
		_, cp, err = s.bithumbApi(code)
		return cp, err
	case m.DomesticETF, m.DomesticGoldETF:
		stock, err := s.kis.DomesticEtfPrice(code)
		return stock.op, err
	case m.ForeignStock, m.ForeignETF:
		_, cp, err := s.kis.ForeignPrice(code)
		return cp, err
	}

	s.lg.Error().Err(err).Msg("Error in ClosingPrice")
	return 0, errors.New("미분류된 종목")
}

const realEstateUrl = "https://www.ep.go.kr/www/contents.do?key=3763"
const realEstateCss = "#contents > table:nth-child(8) > tbody > tr:nth-child(2) > td:nth-child(6)"

func (s *Scraper) RealEstateStatus() (string, error) {
	s.lg.Info().Msg("Starting RealEstateStatus")
	return crawl(realEstateUrl, realEstateCss)
}

const exRateUrl = "https://search.naver.com/search.naver?where=nexearch&sm=top_hty&fbm=0&ie=utf8&query=%ED%99%98%EC%9C%A8"
const exRateCssPath = "#main_pack > section.sc_new.cs_nexchangerate > div:nth-child(1) > div.exchange_bx._exchange_rate_calculator > div > div.inner > div:nth-child(3) > div.num > div > span"

func (s *Scraper) ExchageRate() float64 {
	s.lg.Info().Msg("Starting ExchageRate")
	if s.exchange.Rate != 0 && !s.exchange.Date.Before(time.Now().Add(-3*time.Hour)) {
		return s.exchange.Rate
	}

	rtn, err := crawl(exRateUrl, exRateCssPath)
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
	return s.kis.Index(Nasdaq)
}

func (s *Scraper) Sp500() (float64, error) {
	s.lg.Info().Msg("Starting Sp500")
	return s.kis.Index(Sp500)
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

const (
	highYieldSpreadUrl = "https://fred.stlouisfed.org/graph/fredgraph.csv?id=BAMLH0A0HYM2"
)

func (s *Scraper) HighYieldSpread() (date string, spread float64, err error) {
	s.lg.Info().Msg("Starting HighYieldSpread")

	resp, err := http.Get(highYieldSpreadUrl)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// Read all content (small file, safe to read fully)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	// Split by lines
	lines := bytes.Split(body, []byte("\n"))
	if len(lines) < 2 {
		panic("no data found")
	}

	// Get the last non-empty line (in case of trailing newline)
	var lastLine []byte
	for i := len(lines) - 1; i >= 0; i-- {
		if len(bytes.TrimSpace(lines[i])) > 0 {
			lastLine = lines[i]
			break
		}
	}

	// Split last line by comma
	line := bytes.Split(lastLine, []byte(","))
	if len(line) < 2 {
		panic("invalid last line format")
	}

	date = string(line[0])
	value := string(line[1])

	spread, err = strconv.ParseFloat(value, 64)
	if err != nil {
		return "", 0, err
	}

	return date, spread, nil
}

const (
	sp500ListUrl = "https://en.wikipedia.org/wiki/List_of_S%26P_500_companies"
)

func (s *Scraper) RecentSP500Entries(targetDate string) ([]m.SP500Company, error) {
	s.lg.Info().Msg("Starting SP500List")

	if !regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`).MatchString(targetDate) {
		return nil, errors.New("invalid date format")
	}

	doc, err := crawlDocument(sp500ListUrl)
	if err != nil {
		return nil, err
	}

	var latestEntry []m.SP500Company

	// Find the first table (S&P 500 component stocks table)
	doc.Find("table.wikitable").First().Find("tr").Each(func(i int, row *goquery.Selection) {
		// Skip header row
		if i == 0 {
			return
		}

		cells := row.Find("td")
		if cells.Length() < 8 {
			return
		}

		// Extract data from table cells
		symbol := strings.TrimSpace(cells.Eq(0).Text())
		security := strings.TrimSpace(cells.Eq(1).Text())
		gicsSector := strings.TrimSpace(cells.Eq(2).Text())
		gicsSubIndustry := strings.TrimSpace(cells.Eq(3).Text())
		headquartersLocation := strings.TrimSpace(cells.Eq(4).Text())
		dateAddedStr := strings.TrimSpace(cells.Eq(5).Text())
		cik := strings.TrimSpace(cells.Eq(6).Text())
		founded := strings.TrimSpace(cells.Eq(7).Text())

		// Parse date added
		dateAdded, err := time.Parse("2006-01-02", dateAddedStr)
		if err != nil {
			// Skip if date cannot be parsed
			return
		}

		// Filter entries after targetDate
		if strings.Compare(targetDate, dateAddedStr) < 0 {
			latestEntry = append(latestEntry, m.SP500Company{
				Symbol:                symbol,
				Security:              security,
				GICS_Sector:           gicsSector,
				GICS_Sub_Industry:     gicsSubIndustry,
				Headquarters_Location: headquartersLocation,
				Date_added:            dateAdded,
				CIK:                   cik,
				Founded:               founded,
			})
		}
	})

	return latestEntry, nil
}

// todo. refactor scraper 변경 필요
func (s *Scraper) Buy(category m.Category, code string, qty uint) error {
	s.lg.Info().Msgf("Starting Buy with category: %v, code: %s", category, code)

	switch category {
	case m.ForeignStock, m.ForeignETF:
		err := s.kis.ForeignBuy(code, qty)
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

const upbitNotice = "https://upbit.com/service_center/notice"
const upbitNoticeCssPath = "#UpbitLayout > div.subMain > div > section > article > div.css-tev1mt > table > tbody > tr > td.css-d1s6vu > a > span"

// #UpbitLayout > div.subMain > div > section > article > div.css-tev1mt > table > tbody > tr:nth-child(6) > td.css-d1s6vu > a > span

func (s *Scraper) AirdropEventUpbit() ([]string, []string, error) {

	titles := make([]string, 0)
	urls := make([]string, 0)

	// 1. Body 읽어오기
	doc, err := crawlSpaBody(upbitNotice)
	if err != nil {
		return nil, nil, err
	}

	isVisited := false

	// 2. 공지 타이틀 및 url 추출하기
	doc.Find(upbitNoticeCssPath).Each(func(_ int, s *goquery.Selection) {
		isVisited = true
		title := s.Text()
		matched, err := regexp.MatchString("에어드랍|퀴즈|받아가", title)
		if err != nil {
			return
		}

		if matched {
			titles = append(titles, title)
			urls = append(urls, s.Parent().AttrOr("href", ""))
		}
	})

	if !isVisited {
		// return nil, nil, errors.New("업비트 공지 크롤링 실패 - cssPath 변경 여부 확인 필요")
	}

	return titles, urls, nil
}

const bithumbNotice = "https://feed.bithumb.com/notice"
const bithumbNoticeCssPath = "#__next > div.content > div > div > ul > li > a > span.NoticeContentList_notice-list__inner__aSUqu"

func (s *Scraper) AirdropEventBithumb() ([]string, []string, error) {

	titles := make([]string, 0)
	urls := make([]string, 0)

	// 1. Body 읽어오기
	doc, err := crawlSpaBodyAvoidingClaudFlare(bithumbNotice)
	if err != nil {
		return nil, nil, err
	}

	// 2. 공지 타이틀 및 url 추출하기
	doc.Find(bithumbNoticeCssPath).Each(func(_ int, s *goquery.Selection) {
		title := s.Text()
		matched, err := regexp.MatchString("에어드랍|퀴즈|받아가", title)
		if err != nil {
			return
		}

		if matched {
			titles = append(titles, title)
			urls = append(urls, s.Parent().AttrOr("href", ""))
		}
	})

	return titles, urls, nil
}

func (s *Scraper) StreamCoinOrders(c chan<- m.MyOrder) error {
	if err := s.upbitMyOrders(func(order *UpbitMyOrders) {
		if order.State == "trade" {
			code, _ := strings.CutPrefix(order.Code, "KRW-")
			order.Code = code
			c <- m.MyOrder{
				Code:  order.Code,
				Price: order.Price,
				Count: order.ExecutedVolume,
			}
		}
	}); err != nil {
		return err
	}

	return nil
}

func (s *Scraper) StreamStockOrders(c chan<- m.MyOrder) error {

	if s.kis.wsConn == nil {
		// Step 1: Issue WebSocket approval key
		approvalResp, err := s.kis.IssueWebSocketApprovalKey()
		if err != nil {
			return fmt.Errorf("Failed to issue WebSocket approval key: %w", err)
		}

		// Step 2: Connect to WebSocket
		err = s.kis.ConnectWebSocket(approvalResp.ApprovalKey)
		if err != nil {
			return err
		}
	}
	defer s.kis.CloseWebSocket()

	if err := s.kis.SubscribeMultipleRealTimeExecution(true, true, &RealTimeExecutionCallbacks{
		DomesticCallback: func(kisOrder *RealTimeExecutionNotification) {
			s.lg.Info().Msgf("Received domestic execution notification: %+v", kisOrder)
			if kisOrder.ExecYN == "2" { // 1=Order/Revise/Cancel, 2=Execution
				price, _ := strconv.ParseFloat(kisOrder.ExecPrice, 64)
				count, _ := strconv.ParseFloat(kisOrder.ExecQty, 64)
				if kisOrder.SellBuyDiv == "01" { // 01=Sell, 02=Buy
					count *= -1
				}
				c <- m.MyOrder{
					Code:  kisOrder.StockCode,
					Price: price,
					Count: count,
				}
			}
		},
		OverseasCallback: func(kisOrder *OverseasRealTimeExecutionNotification) {
			s.lg.Info().Msgf("Received overseas execution notification: %+v", kisOrder)
			if kisOrder.ExecYN == "2" { // 1=Order/Revise/Cancel, 2=Execution
				// Insert decimal point at 4th position from the right
				priceStr := kisOrder.ExecPrice
				if len(priceStr) > 4 {
					priceStr = priceStr[:len(priceStr)-4] + "." + priceStr[len(priceStr)-4:]
				}
				price, _ := strconv.ParseFloat(priceStr, 64)
				count, _ := strconv.ParseFloat(kisOrder.ExecQty, 64)
				if kisOrder.SellBuyDiv == "01" { // 01=Sell, 02=Buy
					count *= -1
				}
				var prefix string
				switch kisOrder.OverseasStockDiv {
				case "6":
					prefix = "NAS-"
				case "7":
					prefix = "NYS-"
				case "8":
					prefix = "AMS-"
				}

				c <- m.MyOrder{
					Code:  prefix + kisOrder.StockShortCode, // todo. market 정보 prefix 추가 필요.
					Price: price,
					Count: count,
				}
			}
		}}); err != nil {
		return err
	}

	return nil
}

// func (s *Scraper) StreamOverseasStockOrders(c chan<- m.MyOrder) error {

// 	if s.kis.wsConn == nil {
// 		// Step 1: Issue WebSocket approval key
// 		approvalResp, err := s.kis.IssueWebSocketApprovalKey()
// 		if err != nil {
// 			return fmt.Errorf("Failed to issue WebSocket approval key: %w", err)
// 		}

// 		// Step 2: Connect to WebSocket
// 		err = s.kis.ConnectWebSocket(approvalResp.ApprovalKey)
// 		if err != nil {
// 		}
// 	}
// 	defer s.kis.CloseWebSocket()

// 	if err := s.kis.SubscribeOverseasRealTimeExecution("", func(kisOrder *OverseasRealTimeExecutionNotification) {
// 		price, _ := strconv.ParseFloat(kisOrder.ExecPrice, 64)
// 		count, _ := strconv.ParseFloat(kisOrder.ExecQty, 64)
// 		c <- m.MyOrder{
// 			Code:  kisOrder.StockShortCode,
// 			Price: price,
// 			Count: count,
// 		}
// 	}); err != nil {
// 		return err
// 	}

// 	return nil
// }
