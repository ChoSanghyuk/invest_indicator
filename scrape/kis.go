package scrape

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	kisTokenUrl                 = "https://openapi.koreainvestment.com:9443/oauth2/tokenP"
	kisDomesticStockUrlForm     = "https://openapi.koreainvestment.com:9443/uapi/domestic-stock/v1/quotations/inquire-price?fid_cond_mrkt_div_code=J&fid_input_iscd=%s"
	kisForeignPriceUrlForm      = "https://openapi.koreainvestment.com:9443/uapi/overseas-price/v1/quotations/price?AUTH=&EXCD=%s&SYMB=%s"
	kisIndexUrlForm             = "https://openapi.koreainvestment.com:9443/uapi/overseas-price/v1/quotations/inquire-daily-chartprice?FID_COND_MRKT_DIV_CODE=N&FID_INPUT_ISCD=%s&FID_INPUT_DATE_1=%s&FID_INPUT_DATE_2=%s&FID_PERIOD_DIV_CODE=D"
	kisDomesticEtfPriceUrlForm  = "https://openapi.koreainvestment.com:9443/uapi/etfetn/v1/quotations/inquire-price?fid_cond_mrkt_div_code=J&fid_input_iscd=%s"
	kisForeignDailyPriceUrlForm = "https://openapi.koreainvestment.com:9443/uapi/overseas-price/v1/quotations/dailyprice?EXCD=%s&SYMB=%s&GUBN=0&BYMD=%s&MODP=0"
	kisDomesticStockBuyUrl      = "https://openapi.koreainvestment.com:9443/uapi/domestic-stock/v1/trading/order-cash"
)

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Expired     string `json:"access_token_token_expired"`
}

func (s *Scraper) KisToken() (string, error) {

	if s.kis.accessToken != "" && strings.Compare(s.kis.tokenExpired, time.Now().Format("2006-01-02 15:04:05")) == 1 {
		return s.kis.accessToken, nil
	}

	var token TokenResponse
	err := sendRequest(kisTokenUrl, http.MethodPost, nil, map[string]string{
		"grant_type": "client_credentials",
		"appkey":     s.kis.appKey,
		"appsecret":  s.kis.appSecret,
	}, &token)
	if err != nil {
		return "", err
	}

	s.kis.accessToken = token.AccessToken
	s.kis.tokenExpired = token.Expired

	s.lg.Debug().Str("token", token.AccessToken)
	return token.AccessToken, nil
}

type KIsResp struct {
	Msg    string            `json:"msg1"`
	MsgCd  string            `json:"msg_cd"`
	Output map[string]string `json:"output"` // value가 string 타입으로 넘어오기에 바로 파싱 X
	RtCd   string            `json:"rt_cd"`
}

type StockPrice struct {
	pp float64
	hp float64
	lp float64
	op float64
	ap float64
}

func (s *Scraper) kisDomesticStockPrice(code string) (StockPrice, error) {

	url := fmt.Sprintf(kisDomesticStockUrlForm, code)

	token, err := s.KisToken()
	if err != nil {
		return StockPrice{}, err
	}

	var rtn KIsResp

	header := map[string]string{
		"Content-Type":  "application/json",
		"authorization": "Bearer " + token,
		"appkey":        s.kis.appKey,
		"appsecret":     s.kis.appSecret,
		"tr_id":         "FHKST01010100",
	}

	err = sendRequest(url, http.MethodGet, header, nil, &rtn)
	if err != nil {
		return StockPrice{}, err
	}

	if rtn.RtCd != "0" {
		return StockPrice{}, errors.New("국내 주식현재가 시세 API 실패 코드 반환")
	}

	pp, err := strconv.ParseFloat(rtn.Output["stck_prpr"], 64)
	if err != nil {
		return StockPrice{}, err
	}

	op, err := strconv.ParseFloat(rtn.Output["stck_oprc"], 64) // 시가
	if err != nil {
		return StockPrice{}, err
	}

	ap, err := strconv.ParseFloat(rtn.Output["wghn_avrg_stck_prc"], 64) // 가중 평균 주식 가격
	if err != nil {
		return StockPrice{}, err
	}

	hp, err := strconv.ParseFloat(rtn.Output["w52_hgpr"], 64)
	if err != nil {
		return StockPrice{}, err
	}

	lp, err := strconv.ParseFloat(rtn.Output["w52_lwpr"], 64)
	if err != nil {
		return StockPrice{}, err
	}

	return StockPrice{
		pp: pp,
		op: op,
		hp: hp,
		lp: lp,
		ap: ap,
	}, nil
}

// 해외주식 현재체결가[v1_해외주식-009]
func (s *Scraper) kisForeignPrice(code string) (pp, cp float64, err error) {

	/*
		NYS : 뉴욕
		NAS : 나스닥
	*/
	parmas := strings.Split(code, "-")
	url := fmt.Sprintf(kisForeignPriceUrlForm, parmas[0], parmas[1]) // Nas

	token, err := s.KisToken()
	if err != nil {
		return 0, 0, err
	}

	var rtn KIsResp

	header := map[string]string{
		"Content-Type":  "application/json",
		"authorization": "Bearer " + token,
		"appkey":        s.kis.appKey,
		"appsecret":     s.kis.appSecret,
		"tr_id":         "HHDFS00000300",
	}

	err = sendRequest(url, http.MethodGet, header, nil, &rtn)
	if err != nil {
		return 0, 0, err
	}

	if rtn.RtCd != "0" {
		return 0, 0, errors.New("해외 주식현재가 시세 API 실패 코드 반환")
	}

	pp, err = strconv.ParseFloat(rtn.Output["last"], 64)
	if err != nil {
		return 0, 0, err
	}

	cp, err = strconv.ParseFloat(rtn.Output["base"], 64) // 전일의 종가
	if err != nil {
		return 0, 0, err
	}

	return pp, cp, nil
}

/*
func (s *Scraper) kisNasdaqIndex() (float64, error) {

	today := time.Now().Format("20060102")
	url := fmt.Sprintf(s.t.ApiBaseUrl("KIS_IDX"), today, today)

	token, err := s.KisToken()
	if err != nil {
		return 0, err
	}

	header := map[string]string{
		"Content-Type":  "application/json",
		"authorization": "Bearer " + token,
		"appkey":        s.kis.appKey,
		"appsecret":     s.kis.appSecret,
		"tr_id":         "FHKST03030100",
	}

	type NasdaqResp struct {
		Msg    string `json:"msg1"`
		MsgCd  string `json:"msg_cd"`
		RtCd   string `json:"rt_cd"`
		Output struct {
			PresentPrice string `json:"ovrs_nmix_prpr"`
		} `json:"output1"` // value가 string 타입으로 넘어오기에 바로 파싱 X
	}
	var rtn NasdaqResp //TempResp

	err = sendRequest(url, http.MethodGet, header, nil, &rtn)
	if err != nil {
		return 0, err
	}

	if rtn.RtCd != "0" {
		return 0, errors.New("나스닥 인덱스 API 조회 실패 코드 반환")
	}

	pp, err := strconv.ParseFloat(rtn.Output.PresentPrice, 64)
	if err != nil {
		return 0, err
	}

	return pp, nil
}
*/

// 해외주식 종목/지수/환율기간별시세(일/주/월/년)[v1_해외주식-012]
/*
해당 API로 미국주식 조회 시, 다우30, 나스닥100, S&P500 종목만 조회 가능합니다.
더 많은 미국주식 종목 시세를 이용할 시에는, 해외주식기간별시세 API
*/

type Index string

const (
	Nasdaq = "COMP"
	Sp500  = "SPX"
)

func (s *Scraper) kisIndex(idx Index) (float64, error) { // todo 반복 코드 모듈화

	today := time.Now().Format("20060102")
	url := fmt.Sprintf(kisIndexUrlForm, idx, today, today)

	token, err := s.KisToken()
	if err != nil {
		return 0, err
	}

	header := map[string]string{
		"Content-Type":  "application/json",
		"authorization": "Bearer " + token,
		"appkey":        s.kis.appKey,
		"appsecret":     s.kis.appSecret,
		"tr_id":         "FHKST03030100",
	}

	type NasdaqResp struct {
		Msg    string `json:"msg1"`
		MsgCd  string `json:"msg_cd"`
		RtCd   string `json:"rt_cd"`
		Output struct {
			PresentPrice string `json:"ovrs_nmix_prpr"`
		} `json:"output1"` // value가 string 타입으로 넘어오기에 바로 파싱 X
	}
	var rtn NasdaqResp //TempResp

	err = sendRequest(url, http.MethodGet, header, nil, &rtn)
	if err != nil {
		return 0, err
	}

	if rtn.RtCd != "0" {
		fmt.Printf("%+v\n", rtn)
		return 0, errors.New("나스닥 인덱스 API 조회 실패 코드 반환")
	}

	pp, err := strconv.ParseFloat(rtn.Output.PresentPrice, 64)
	if err != nil {
		return 0, err
	}

	return pp, nil
}

func (s *Scraper) kisDomesticEtfPrice(code string) (StockPrice, error) {

	url := fmt.Sprintf(kisDomesticEtfPriceUrlForm, code)

	token, err := s.KisToken()
	if err != nil {
		return StockPrice{}, err
	}

	var rtn KIsResp

	header := map[string]string{
		"Content-Type":  "application/json",
		"authorization": "Bearer " + token,
		"appkey":        s.kis.appKey,
		"appsecret":     s.kis.appSecret,
		"tr_id":         "FHPST02400000",
	}

	err = sendRequest(url, http.MethodGet, header, nil, &rtn)
	if err != nil {
		return StockPrice{}, err
	}

	if rtn.RtCd != "0" {
		return StockPrice{}, errors.New("국내 주식현재가 시세 API 실패 코드 반환")
	}

	pp, err := strconv.ParseFloat(rtn.Output["stck_prpr"], 64)
	if err != nil {
		return StockPrice{}, err
	}

	op, err := strconv.ParseFloat(rtn.Output["stck_oprc"], 64) // 시가
	if err != nil {
		return StockPrice{}, err
	}

	hp, err := strconv.ParseFloat(rtn.Output["stck_dryy_hgpr"], 64) // 연중 최고가
	if err != nil {
		return StockPrice{}, err
	}

	lp, err := strconv.ParseFloat(rtn.Output["stck_dryy_lwpr"], 64) // 연중 최저가
	if err != nil {
		return StockPrice{}, err
	}

	return StockPrice{
		pp: pp,
		op: op,
		hp: hp,
		lp: lp,
	}, nil
}

type Output2 struct {
	Date  string `json:"xymd"`
	Price string `json:"clos"`
}

// todo
type KisPeriodResp struct {
	Msg     string            `json:"msg1"`
	MsgCd   string            `json:"msg_cd"`
	Output  map[string]string `json:"output1"` // value가 string 타입으로 넘어오기에 바로 파싱 X
	Output2 []Output2         `json:"output2"` // value가 string 타입으로 넘어오기에 바로 파싱 X
	RtCd    string            `json:"rt_cd"`
}

func (s *Scraper) kisForeignAvg(code string) (ap float64, n int, err error) {
	sum := 0.0
	day := time.Now().Format("20060102")

loop:

	parmas := strings.Split(code, "-")
	url := fmt.Sprintf(kisForeignDailyPriceUrlForm, parmas[0], parmas[1], day) // Nas

	token, err := s.KisToken()
	if err != nil {
		return 0, 0, err
	}

	var rtn KisPeriodResp

	header := map[string]string{
		"Content-Type":  "application/json",
		"authorization": "Bearer " + token,
		"appkey":        s.kis.appKey,
		"appsecret":     s.kis.appSecret,
		"tr_id":         "HHDFS76240000",
	}

	err = sendRequest(url, http.MethodGet, header, nil, &rtn)
	if err != nil {
		return 0, 0, err
	}

	if rtn.RtCd != "0" {
		return 0, 0, errors.New("해외 주식기간별 시세 API 실패 코드 반환")
	}

	if len(rtn.Output2) == 0 {
		return 0, 0, errors.New("해외 주식기간별 시세 API 기간 데이터 없음")
	}

	for _, r := range rtn.Output2 {
		p, err := strconv.ParseFloat(r.Price, 64) // 연중 최고가
		if err != nil {
			return 0, 0, err
		}
		sum += p
	}
	n += len(rtn.Output2)

	if n == 100 {
		parsedDay, _ := time.Parse("20060102", rtn.Output2[n-1].Date)
		day = parsedDay.AddDate(0, 0, -1).Format("20060102")
		goto loop
	}

	x := sum / float64(n)

	return math.Round(x*100) / 100, n, nil
}

/*
국내주식주문 매도 : TTTC0011U
국내주식주문 매수 : TTTC0012U
*/
func (s *Scraper) kisDomesticBuy(code string, qty uint) error {

	token, err := s.KisToken()
	if err != nil {
		return err
	}

	var rtn KIsResp

	header := map[string]string{
		"Content-Type":  "application/json",
		"authorization": "Bearer " + token,
		"appkey":        s.kis.appKey,
		"appsecret":     s.kis.appSecret,
		"tr_id":         "TTTC0012U", // 국내 매수
		"custtype":      "P",
	}

	accounts := strings.Split(s.kis.account, "-")
	body := map[string]string{
		"CANO":         accounts[0],            // 종합계좌번호	String	Y	8	계좌번호 체계(8-2)의 앞 8자리
		"ACNT_PRDT_CD": accounts[1],            // 계좌상품코드	String	Y	2	계좌번호 체계(8-2)의 뒤 2자리
		"PDNO":         code,                   // 상품번호	String	Y	12	종목코드
		"ORD_DVSN":     "01",                   // 00 : 지정가 / 01 : 시장가
		"ORD_QTY":      fmt.Sprintf("%d", qty), // 주문수량
		"ORD_UNPR":     "0",                    // 주문단가	String	Y	31	1주당 가격. 시장가 = 0
	}

	err = sendRequest(kisDomesticStockBuyUrl, http.MethodPost, header, body, &rtn)
	if err != nil {
		return err
	}

	if rtn.RtCd != "0" {
		s.lg.Error().Any("response", rtn).Msg("국내 주식 구매 실패")
		return errors.New("해외 주식 거래 API 실패 코드 반환")
	}
	return nil
}

func (s *Scraper) kisForeignBuy(code string, qty uint) error {

	url := ""
	token, err := s.KisToken()
	if err != nil {
		return err
	}

	var rtn KIsResp

	header := map[string]string{
		"Content-Type":  "application/json",
		"authorization": "Bearer " + token,
		"appkey":        s.kis.appKey,
		"appsecret":     s.kis.appSecret,
		"tr_id":         "TTTT1002U", // 미국 매수
	}

	accounts := strings.Split(s.kis.account, "-")

	ovrsExcgCd := ""
	if strings.HasPrefix(code, "NAS") {
		ovrsExcgCd = "NASD"
	} else if strings.HasPrefix(code, "NYS") {
		ovrsExcgCd = "NYSE"
	}

	body := map[string]string{
		"CANO":            accounts[0],            // 종합계좌번호	String	Y	8	계좌번호 체계(8-2)의 앞 8자리
		"ACNT_PRDT_CD":    accounts[1],            // 계좌상품코드	String	Y	2	계좌번호 체계(8-2)의 뒤 2자리
		"OVRS_EXCG_CD":    ovrsExcgCd,             // 해외거래소코드	String	Y	4. NASD : 나스닥 / NYSE : 뉴욕
		"PDNO":            code,                   // 상품번호	String	Y	12	종목코드
		"ORD_QTY":         fmt.Sprintf("%d", qty), // 주문수량
		"OVRS_ORD_UNPR":   "0",                    // 해외주문단가	String	Y	31	1주당 가격. 시장가 = 0
		"ORD_SVR_DVSN_CD": "0",                    // 주문서버구분코드. 0 고정.
		"ORD_DVSN":        "00",                   // 지정가. 시장가 코드가 없음. 주문단가가 0이면 시장가로 되는건가.
	}

	err = sendRequest(url, http.MethodPost, header, body, &rtn)
	if err != nil {
		return err
	}

	if rtn.RtCd != "0" {
		s.lg.Error().Any("response", rtn).Msg("해외 주식 구매 실패")
		return errors.New("해외 주식 거래 API 실패 코드 반환")
	}

	return nil
}
