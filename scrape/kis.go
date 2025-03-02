package scrape

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
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

	url := "https://openapi.koreainvestment.com:9443/oauth2/tokenP"

	var token TokenResponse
	err := sendRequest(url, http.MethodPost, nil, map[string]string{
		"grant_type": "client_credentials",
		"appkey":     s.kis.appKey,
		"appsecret":  s.kis.appSecret,
	}, &token)
	if err != nil {
		return "", err
	}

	s.kis.accessToken = token.AccessToken
	s.kis.tokenExpired = token.Expired

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

	url := s.t.ApiBaseUrl("KIS")
	if url == "" {
		return StockPrice{}, errors.New("URL 미존재")
	}
	url = fmt.Sprintf(url, code)

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
// https://openapi.koreainvestment.com:9443/uapi/overseas-price/v1/quotations/price?AUTH=""&EXCD=%s&SYMB=%s
func (s *Scraper) kisForeignPrice(code string) (pp, cp float64, err error) {

	url := s.t.ApiBaseUrl("KIS_FOR")
	if url == "" {
		return 0, 0, errors.New("URL 미존재")
	}
	/*
		NYS : 뉴욕
		NAS : 나스닥
	*/
	parmas := strings.Split(code, "-")
	url = fmt.Sprintf(url, parmas[0], parmas[1]) // Nas

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

// todo. transmitter 필요한가??
func (s *Scraper) kisIndex(idx Index) (float64, error) { // todo 반복 코드 모듈화

	today := time.Now().Format("20060102")
	// COMP
	base := "https://openapi.koreainvestment.com:9443/uapi/overseas-price/v1/quotations/inquire-daily-chartprice?FID_COND_MRKT_DIV_CODE=N&FID_INPUT_ISCD=%s&FID_INPUT_DATE_1=%s&FID_INPUT_DATE_2=%s&FID_PERIOD_DIV_CODE=D"
	url := fmt.Sprintf(base, idx, today, today) // 여기 url에서 sp500 파람 전달

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

func (s *Scraper) kisDomesticEtfPrice(code string) (StockPrice, error) {

	url := s.t.ApiBaseUrl("KIS_ETF")
	if url == "" {
		return StockPrice{}, errors.New("URL 미존재")
	}
	url = fmt.Sprintf(url, code)

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

	body := map[string]string{
		"CANO":            "",                     // 종합계좌번호	String	Y	8	계좌번호 체계(8-2)의 앞 8자리
		"ACNT_PRDT_CD":    "",                     // 계좌상품코드	String	Y	2	계좌번호 체계(8-2)의 뒤 2자리
		"OVRS_EXCG_CD":    "",                     // 해외거래소코드	String	Y	4. NASD : 나스닥 / NYSE : 뉴욕
		"PDNO":            code,                   // 상품번호	String	Y	12	종목코드
		"ORD_QTY":         fmt.Sprintf("%d", qty), // 주문수량
		"OVRS_ORD_UNPR":   "0",                    // 해외주문단가	String	Y	31	1주당 가격. 시장가 = 0
		"ORD_SVR_DVSN_CD": "0",                    //주문서버구분코드
		"ORD_DVSN":        "00",                   // 지정가
	}

	err = sendRequest(url, http.MethodPost, header, body, &rtn)
	if err != nil {
		return err
	}

	if rtn.RtCd != "0" {
		fmt.Printf("%v", rtn)
		return errors.New("해외 주식 거래 API 실패 코드 반환")
	}
	return nil
}
