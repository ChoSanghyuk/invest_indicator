package scrape

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

const (
	ProdBaseURL = "https://openapi.koreainvestment.com:9443"
	MockBaseURL = "https://openapivts.koreainvestment.com:29443"
)

// Kis handles all Korea Investment & Securities API operations
type Kis struct {
	appKey       string
	appSecret    string
	accessToken  string
	tokenExpired string
	account      string
	isMock       bool // true for mock/test environment
	lg           zerolog.Logger

	// WebSocket fields
	wsConn        *websocket.Conn
	wsMutex       sync.Mutex
	wsAESKey      string
	wsAESIV       string
	wsApprovalKey string
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Expired     string `json:"access_token_token_expired"`
}

// NewKis creates a new Kis instance
func NewKis(appKey, appSecret, account string) *Kis {
	return &Kis{
		appKey:    appKey,
		appSecret: appSecret,
		account:   account,
		lg:        zerolog.New(os.Stdout).With().Str("Module", "Kis").Timestamp().Logger(),
	}
}

// SetAccessToken sets the access token and expiration time
func (k *Kis) SetAccessToken(token string) {
	k.accessToken = token
	k.tokenExpired = time.Now().Add(time.Duration(1) * time.Hour).Format("2006-01-02 15:04:05")
}

func (k *Kis) KisToken() (string, error) {

	endpoint := "/oauth2/tokenP"
	url := k.getBaseURL() + endpoint

	if k.accessToken != "" && strings.Compare(k.tokenExpired, time.Now().Format("2006-01-02 15:04:05")) == 1 {
		return k.accessToken, nil
	}

	var token TokenResponse
	err := sendRequest(url, http.MethodPost, nil, map[string]string{
		"grant_type": "client_credentials",
		"appkey":     k.appKey,
		"appsecret":  k.appSecret,
	}, &token)
	if err != nil {
		return "", err
	}

	k.accessToken = token.AccessToken
	k.tokenExpired = token.Expired

	k.lg.Debug().Str("token", token.AccessToken)
	return token.AccessToken, nil
}

func (k *Kis) DomesticStockPrice(code string) (StockPrice, error) {

	endpoint := "/uapi/domestic-stock/v1/quotations/inquire-price"
	url := k.getBaseURL() + endpoint

	queryParams := map[string]string{
		"fid_input_iscd":         code,
		"fid_cond_mrkt_div_code": "J",
	}
	var rtn KisResp

	err := k.executeGetRequest(url, "FHKST01010100", queryParams, &rtn)
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
func (k *Kis) ForeignPrice(code string) (pp, cp float64, err error) {

	endpoint := "/uapi/overseas-price/v1/quotations/price"
	url := k.getBaseURL() + endpoint

	/*
		NYS : 뉴욕
		NAS : 나스닥
	*/
	parmas := strings.Split(code, "-")
	queryParams := map[string]string{
		"AUTH": "",
		"EXCD": parmas[0],
		"SYMB": parmas[1],
	}
	var rtn KisResp

	err = k.executeGetRequest(url, "HHDFS00000300", queryParams, &rtn)
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

func (k *Kis) Index(idx Index) (float64, error) { // todo 반복 코드 모듈화

	endpoint := "/uapi/overseas-price/v1/quotations/inquire-daily-chartprice"
	url := k.getBaseURL() + endpoint

	today := time.Now().Format("20060102")

	queryParams := map[string]string{
		"FID_COND_MRKT_DIV_CODE": "N",
		"FID_INPUT_ISCD":         string(idx),
		"FID_INPUT_DATE_1":       today,
		"FID_INPUT_DATE_2":       today,
		"FID_PERIOD_DIV_CODE":    "D",
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

	err := k.executeGetRequest(url, "FHKST03030100", queryParams, &rtn)
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

func (k *Kis) DomesticEtfPrice(code string) (StockPrice, error) {

	endpoint := "/uapi/etfetn/v1/quotations/inquire-price"
	url := k.getBaseURL() + endpoint

	queryParams := map[string]string{
		"fid_cond_mrkt_div_code": "J",
		"fid_input_iscd":         code,
	}

	var rtn KisResp

	err := k.executeGetRequest(url, "FHPST02400000", queryParams, &rtn)

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

func (k *Kis) ForeignAvg(code string) (ap float64, n int, err error) {
	endpoint := "/uapi/overseas-price/v1/quotations/dailyprice"
	url := k.getBaseURL() + endpoint

	sum := 0.0
	day := time.Now().Format("20060102")

loop:

	parmas := strings.Split(code, "-")
	queryParams := map[string]string{
		"EXCD": parmas[0],
		"SYMB": parmas[1],
		"GUBN": "0",
		"BYMD": day,
		"MODP": "0",
	}
	var rtn KisPeriodResp

	err = k.executeGetRequest(url, "HHDFS76240000", queryParams, &rtn)

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

func (k *Kis) InquireDailyCcld(startDate string, endDate string, stockCode string) (*InquireDailyCcldResponse, error) {
	k.lg.Debug().
		Str("startDate", startDate).
		Str("endDate", endDate).
		Str("stockCode", stockCode).
		Msg("InquireDailyCcld called")

	accounts := strings.Split(k.account, "-") // Build query parameters
	queryParams := map[string]string{
		"CANO":            accounts[0],
		"ACNT_PRDT_CD":    accounts[1],
		"INQR_STRT_DT":    startDate, // Start date (YYYYMMDD)
		"INQR_END_DT":     endDate,   // End date (YYYYMMDD)
		"SLL_BUY_DVSN_CD": "00",      // 00: all, 01: sell, 02: buy
		"PDNO":            stockCode, // Stock code (empty for all)
		"CCLD_DVSN":       "00",      // 00: all, 01: executed, 02: unexecuted
		"ORD_GNO_BRNO":    "",        // Order general branch number (empty for all)
		"ODNO":            "",        // Order number (empty for all)
		"INQR_DVSN":       "01",      // 00: 역순, 01: 정순
		"INQR_DVSN_1":     "",        // Inquiry division 1
		"INQR_DVSN_3":     "00",      // Inquiry division 3
		"EXCG_ID_DVSN_CD": "KRX",
		"CTX_AREA_FK100":  "", // Continuation key (empty for first call)
		"CTX_AREA_NK100":  "", // Continuation key (empty for first call)
	}

	// Determine transaction ID based on environment
	trId := "TTTC0081R" // Real environment

	resp, err := k.executeInquireDailyCcld(queryParams, trId)
	if err != nil {
		k.lg.Error().Err(err).Msg("InquireDailyCcld failed")
		return nil, err
	}

	k.lg.Info().
		Int("recordCount", len(resp.Output)).
		Msg("InquireDailyCcld succeeded")
	return resp, nil
}

// executeInquireDailyCcld executes the daily order execution inquiry API call
func (k *Kis) executeInquireDailyCcld(queryParams map[string]string, trId string) (*InquireDailyCcldResponse, error) {

	endpoint := "/uapi/domestic-stock/v1/trading/inquire-daily-ccld"
	url := k.getBaseURL() + endpoint

	var resp InquireDailyCcldResponse
	if err := k.executeGetRequest(url, trId, queryParams, &resp); err != nil {
		return nil, err
	}

	// Check response code
	if resp.RTCd != "0" {
		return nil, fmt.Errorf("API error: code=%s, msg=%s", resp.MsgCd, resp.Msg1)
	}

	return &resp, nil
}

// InquireDailyCcldResponse represents the response for daily order execution inquiry
type InquireDailyCcldResponse struct {
	RTCd         string                           `json:"rt_cd"`  // Success/failure code
	MsgCd        string                           `json:"msg_cd"` // Message code
	Msg1         string                           `json:"msg1"`   // Message
	CtxAreaFk100 string                           `json:"ctx_area_fk100"`
	CtxAreaNk100 string                           `json:"ctx_area_nk100"`
	Output       []InquireDailyCcldResponseOutput `json:"output"` // Output data array
}

// InquireDailyCcldResponseOutput represents a single daily order execution record
type InquireDailyCcldResponseOutput struct {
	OrdDt        string `json:"ord_dt"`               // Order date (YYYYMMDD)
	OrdGnoNo     string `json:"ord_gno_no"`           // Order general number
	OrdNo        string `json:"orgn_odno"`            // Original order number
	OrdTmd       string `json:"ord_tmd"`              // Order time
	PdNo         string `json:"pdno"`                 // Stock code
	PdNm         string `json:"prdt_name"`            // Product name
	SllBuyDvsnCd string `json:"sll_buy_dvsn_cd"`      // Sell/buy division code (01: sell, 02: buy)
	SllBuyDvsnNm string `json:"sll_buy_dvsn_cd_name"` // Sell/buy division name
	OrdDvsnCd    string `json:"ord_dvsn_cd"`          // Order division code
	OrdDvsnNm    string `json:"ord_dvsn_name"`        // Order division name
	OrdQty       string `json:"ord_qty"`              // Order quantity
	OrdUnpr      string `json:"ord_unpr"`             // Order unit price
	OrdTamt      string `json:"ord_tamt"`             // Order total amount
	CcldQty      string `json:"tot_ccld_qty"`         // Total executed quantity
	AvgPrvs      string `json:"avg_prvs"`             // Average execution price
	CcldAmt      string `json:"tot_ccld_amt"`         // Total executed amount
	RejtQty      string `json:"rmn_qty"`              // Remaining quantity
	RejtRsn      string `json:"rejt_rsn"`             // Rejection reason
	OrdDvsnName  string `json:"ord_dvsn_cd_name"`     // Order division code name
}

// kisDomesticBuy, kisForeignBuy executeRequestWithMethod 소스 적용 필요
/*
국내주식주문 매도 : TTTC0011U
국내주식주문 매수 : TTTC0012U
*/
func (k *Kis) DomesticBuy(code string, qty uint) error {

	endpoint := "/uapi/domestic-stock/v1/trading/order-cash"
	url := k.getBaseURL() + endpoint

	token, err := k.KisToken()
	if err != nil {
		return err
	}

	var rtn KisResp

	header := map[string]string{
		"Content-Type":  "application/json",
		"authorization": "Bearer " + token, //
		"appkey":        k.appKey,
		"appsecret":     k.appSecret,
		"tr_id":         "TTTC0012U", // 국내 매수
		"custtype":      "P",
	}

	accounts := strings.Split(k.account, "-")
	body := map[string]string{
		"CANO":         accounts[0],            // 종합계좌번호	String	Y	8	계좌번호 체계(8-2)의 앞 8자리
		"ACNT_PRDT_CD": accounts[1],            // 계좌상품코드	String	Y	2	계좌번호 체계(8-2)의 뒤 2자리
		"PDNO":         code,                   // 상품번호	String	Y	12	종목코드
		"ORD_DVSN":     "01",                   // 00 : 지정가 / 01 : 시장가
		"ORD_QTY":      fmt.Sprintf("%d", qty), // 주문수량
		"ORD_UNPR":     "0",                    // 주문단가	String	Y	31	1주당 가격. 시장가 = 0
	}

	err = sendRequest(url, http.MethodPost, header, body, &rtn)
	if err != nil {
		return err
	}

	if rtn.RtCd != "0" {
		k.lg.Error().Any("response", rtn).Msg("국내 주식 구매 실패")
		return errors.New("해외 주식 거래 API 실패 코드 반환")
	}
	return nil
}

func (k *Kis) ForeignBuy(code string, qty uint) error {

	endpoint := "/uapi/overseas-stock/v1/trading/order"
	url := k.getBaseURL() + endpoint

	token, err := k.KisToken()
	if err != nil {
		return err
	}

	var rtn KisResp

	header := map[string]string{
		"Content-Type":  "application/json",
		"authorization": "Bearer " + token,
		"appkey":        k.appKey,
		"appsecret":     k.appSecret,
		"tr_id":         "TTTT1002U", // 미국 매수
	}

	accounts := strings.Split(k.account, "-")

	ovrsExcgCd := ""
	switch strings.Split(code, "-")[0] {
	case "NAS":
		ovrsExcgCd = "NASD"
	case "NYS":
		ovrsExcgCd = "NYSE"
	case "AMS":
		ovrsExcgCd = "AMEX"
	default:
		return errors.New("미존재 거래소 코드")
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
		k.lg.Error().Any("response", rtn).Msg("해외 주식 구매 실패")
		return errors.New("해외 주식 거래 API 실패 코드 반환")
	}

	return nil
}

// executeRequest is a common method to execute HTTP POST requests to KIS API
func (k *Kis) executeRequest(endpoint string, trId string, requestBody interface{}, responseBody interface{}) error {
	return k.executeRequestWithMethod("POST", endpoint, trId, requestBody, nil, responseBody)
}

// executeGetRequest is a common method to execute HTTP GET requests to KIS API with query parameters
func (k *Kis) executeGetRequest(endpoint string, trId string, queryParams map[string]string, responseBody interface{}) error {
	return k.executeRequestWithMethod("GET", endpoint, trId, nil, queryParams, responseBody)
}

func (k *Kis) executeRequestWithMethod(method string, url string, trId string, requestBody interface{}, queryParams map[string]string, responseBody interface{}) error {

	// Add query parameters for GET requests
	if method == "GET" && len(queryParams) > 0 {
		params := ""
		for key, value := range queryParams {
			if params != "" {
				params += "&"
			}
			params += key + "=" + value
		}
		url += "?" + params
	}

	k.lg.Debug().
		Str("method", method).
		Str("url", url).
		Str("trId", trId).
		Msg("Executing API request")

	var httpReq *http.Request
	var err error

	if method == "POST" && requestBody != nil {
		reqBody, err := json.Marshal(requestBody)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		httpReq, err = http.NewRequest(method, url, bytes.NewBuffer(reqBody))
	} else {
		httpReq, err = http.NewRequest(method, url, nil)
	}

	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	token, err := k.KisToken()
	if err != nil {
		return err
	}

	// Set required headers
	httpReq.Header.Set("Content-Type", "application/json; charset=utf-8")
	httpReq.Header.Set("authorization", "Bearer "+token)
	httpReq.Header.Set("appkey", k.appKey)
	httpReq.Header.Set("appsecret", k.appSecret)
	httpReq.Header.Set("tr_id", trId)
	httpReq.Header.Set("custtype", "P")

	client := &http.Client{}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	k.lg.Debug().
		Int("status", httpResp.StatusCode).
		Str("body", string(body)).
		Msg("Response received")

	if httpResp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP request failed with status %d: %s", httpResp.StatusCode, string(body))
	}

	if err := json.Unmarshal(body, responseBody); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return nil
}

// getBaseURL returns the appropriate base URL based on environment
func (k *Kis) getBaseURL() string {
	if k.isMock {
		return MockBaseURL
	}
	return ProdBaseURL
}
