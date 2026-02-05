package scrape

type KisResp struct {
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

// WebSocketApprovalKeyRequest represents the request for WebSocket approval key issuance
type WebSocketApprovalKeyRequest struct {
	GrantType string `json:"grant_type" validate:"required"` // Must be "client_credentials"
	AppKey    string `json:"appkey" validate:"required"`     // App key from Korea Investment Securities
	SecretKey string `json:"secretkey" validate:"required"`  // Secret key from Korea Investment Securities
}

// WebSocketApprovalKeyResponse represents the response for WebSocket approval key issuance
type WebSocketApprovalKeyResponse struct {
	ApprovalKey string `json:"approval_key"` // WebSocket approval key (valid for 24 hours)
}

// WebSocketSubscribeRequest represents the WebSocket subscription request
type WebSocketSubscribeRequest struct {
	Header WebSocketSubscribeRequestHeader `json:"header"`
	Body   WebSocketSubscribeRequestBody   `json:"body"`
}

// WebSocketSubscribeRequestHeader represents the header for WebSocket subscription
type WebSocketSubscribeRequestHeader struct {
	ApprovalKey string `json:"approval_key" validate:"required"` // WebSocket approval key
	CustType    string `json:"custtype" validate:"required"`     // Customer type: P (individual) or B (corporate)
	TrType      string `json:"tr_type" validate:"required"`      // Transaction type: 1 (subscribe) or 2 (unsubscribe)
	ContentType string `json:"content-type"`                     // Content type (utf-8)
}

// WebSocketSubscribeRequestBody represents the body for WebSocket subscription
type WebSocketSubscribeRequestBody struct {
	Input WebSocketSubscribeRequestInput `json:"input"`
}

// WebSocketSubscribeRequestInput represents the input for WebSocket subscription
type WebSocketSubscribeRequestInput struct {
	TrID  string `json:"tr_id" validate:"required"`  // Transaction ID (e.g., H0STCNI0)
	TrKey string `json:"tr_key" validate:"required"` // HTS ID
}

// WebSocketSubscribeResponse represents the WebSocket subscription response
type WebSocketSubscribeResponse struct {
	Header WebSocketSubscribeResponseHeader `json:"header"`
	Body   WebSocketSubscribeResponseBody   `json:"body"`
}

// WebSocketSubscribeResponseHeader represents the response header
type WebSocketSubscribeResponseHeader struct {
	TrID    string `json:"tr_id"`
	TrKey   string `json:"tr_key"`
	Encrypt string `json:"encrypt"`
}

// WebSocketSubscribeResponseBody represents the response body
type WebSocketSubscribeResponseBody struct {
	RtCd   string                           `json:"rt_cd"`
	MsgCd  string                           `json:"msg_cd"`
	Msg1   string                           `json:"msg1"`
	Output WebSocketSubscribeResponseOutput `json:"output"`
}

// WebSocketSubscribeResponseOutput contains encryption keys for real-time data
type WebSocketSubscribeResponseOutput struct {
	IV  string `json:"iv"`  // AES256 IV for decryption
	Key string `json:"key"` // AES256 Key for decryption
}

// RealTimeExecutionNotification represents a real-time execution notification
type RealTimeExecutionNotification struct {
	CustID         string // Customer ID (고객 ID)
	AcctNo         string // Account number (계좌번호)
	OrderNo        string // Order number (주문번호)
	OrigOrderNo    string // Original order number (원주문번호)
	SellBuyDiv     string // Sell/Buy division (매도매수구분) 01:Sell, 02:Buy
	ReviseDiv      string // Revise division (정정구분) 0:Normal, 1:Revise, 2:Cancel
	OrderKind      string // Order kind (주문종류)
	OrderCond      string // Order condition (주문조건) 0:None, 1:IOC, 2:FOK
	StockCode      string // Stock code (주식 단축 종목코드)
	ExecQty        string // Execution quantity (체결 수량)
	ExecPrice      string // Execution price (체결단가)
	StockExecTime  string // Stock execution time (주식 체결 시간)
	RefuseYN       string // Refuse flag (거부여부) 0:Approved, 1:Refused
	ExecYN         string // Execution flag (체결여부) 1:Order/Revise/Cancel/Refuse, 2:Execution
	AcceptYN       string // Accept flag (접수여부) 1:Order accepted, 2:Confirmed, 3:Cancel(FOK/IOC)
	BranchNo       string // Branch number (지점번호)
	OrderQty       string // Order quantity (주문수량)
	AcctName       string // Account name (계좌명)
	OrderCondPrice string // Order condition price (호가조건가격)
	OrderExchDiv   string // Order exchange division (주문거래소 구분) 1:KRX, 2:NXT, 3:SOR-KRX, 4:SOR-NXT
	PopupYN        string // Popup display flag (실시간체결창 표시여부)
	Filler         string // Filler (필러)
	CreditDiv      string // Credit division (신용구분)
	CreditLoanDate string // Credit loan date (신용대출일자)
	ExecStockName  string // Execution stock name (체결종목명)
	OrderPrice     string // Order price (주문가격)
}

// OverseasRealTimeExecutionNotification represents an overseas stock real-time execution notification
type OverseasRealTimeExecutionNotification struct {
	CustID             string // Customer ID (고객 ID)
	AcctNo             string // Account number (계좌번호)
	OrderNo            string // Order number (주문번호)
	OrigOrderNo        string // Original order number (원주문번호)
	SellBuyDiv         string // Sell/Buy division (매도매수구분) 01:Sell, 02:Buy, 03:Full Sell, 04:Return Buy
	ReviseDiv          string // Revise division (정정구분) 0:Normal, 1:Revise, 2:Cancel
	OrderKind2         string // Order kind 2 (주문종류2) 1:Market, 2:Limit, 6:Fractional Market, 7:Fractional Limit, A:MOO, B:LOO, C:MOC, D:LOC
	StockShortCode     string // Stock short code (주식 단축 종목코드)
	ExecQty            string // Execution quantity (체결수량) - For order notification: order quantity, For execution notification: execution quantity
	ExecPrice          string // Execution price (체결단가) - Decimal point position varies by country (US:4, JP:1, CN:3, HK:3, VN:0)
	StockExecTime      string // Stock execution time (주식 체결 시간) - Not available for some exchanges, use timestamp when received
	RefuseYN           string // Refuse flag (거부여부) 0:Approved, 1:Refused
	ExecYN             string // Execution flag (체결여부) 1:Order/Revise/Cancel/Refuse, 2:Execution
	AcceptYN           string // Accept flag (접수여부) 1:Order accepted, 2:Confirmed, 3:Cancel(FOK/IOC)
	BranchNo           string // Branch number (지점번호)
	OrderQty           string // Order quantity (주문 수량) - For order notification: not output, For execution notification: order quantity
	AcctName           string // Account name (계좌명)
	ExecStockName      string // Execution stock name (체결종목명)
	OverseasStockDiv   string // Overseas stock division (해외종목구분) 4:HK(HKD), 5:ShangB(USD), 6:NASDAQ, 7:NYSE, 8:AMEX, 9:OTCB, C:HK(CNY), A:ShangA(CNY), B:ShenzB(HKD), D:Tokyo, E:Hanoi, F:HCMC
	CollateralTypeCode string // Collateral type code (담보유형코드) 10:Cash, 15:Overseas stock collateral loan
	CollateralLoanDate string // Collateral loan date (담보대출일자) - Loan date (YYYYMMDD)
	SplitBuyStartTm    string // Split buy/sell start time (분할매수/매도 시작시간) - HHMMSS
	SplitBuyEndTm      string // Split buy/sell end time (분할매수/매도 종료시간) - HHMMSS
	TimeDivType        string // Time division type (시간분할타입유형) 00:Direct time setting, 02:Until regular session
	ExecPrice12        string // Execution price 12 (체결단가12)
}
