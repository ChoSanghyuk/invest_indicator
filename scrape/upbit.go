package scrape

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

func WithUpbitToken(accessKey string, secretKey string) Option {

	payload := jwt.MapClaims{
		"access_key": accessKey,
		"nonce":      uuid.New().String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	jwtToken, err := token.SignedString([]byte(secretKey))

	return func(s *Scraper) error {
		if err != nil {
			return err
		}
		s.upbit.token = jwtToken
		return nil
	}
}

const upbitUrlForm = "https://api.upbit.com/v1/candles/days?market=%s&count=1"

func (s Scraper) upbitApi(sym string) (float64, float64, error) {

	url := fmt.Sprintf(upbitUrlForm, "KRW-"+sym) // KRW-SYMBOL 형식

	var rtn []map[string]any
	err := sendRequest(url, http.MethodGet, nil, nil, &rtn)
	if err != nil {
		return 0, 0, err
	}

	return rtn[0]["trade_price"].(float64), rtn[0]["opening_price"].(float64), nil // 시가 = 전날 종가
}

type UpbitMyOrders struct {
	Type            string  `json:"type"`
	Code            string  `json:"code"`
	UUID            string  `json:"uuid"`
	AskBid          string  `json:"ask_bid"`
	OrderType       string  `json:"order_type"`
	State           string  `json:"state"`
	TradeUUID       string  `json:"trade_uuid"`
	Price           float64 `json:"price"`
	AvgPrice        float64 `json:"avg_price"`
	Volume          float64 `json:"volume"`
	RemainingVolume float64 `json:"remaining_volume"`
	ExecutedVolume  float64 `json:"executed_volume"`
	TradesCount     int     `json:"trades_count"`
	ReservedFee     float64 `json:"reserved_fee"`
	RemainingFee    float64 `json:"remaining_fee"`
	PaidFee         float64 `json:"paid_fee"`
	Locked          float64 `json:"locked"`
	ExecutedFunds   float64 `json:"executed_funds"`
	TimeInForce     string  `json:"time_in_force"`
	TradeFee        float64 `json:"trade_fee"`
	IsMaker         bool    `json:"is_maker"`
	Identifier      string  `json:"identifier"`
	SMPType         string  `json:"smp_type"`
	PreventedVolume float64 `json:"prevented_volume"`
	PreventedLocked float64 `json:"prevented_locked"`
	TradeTimestamp  int64   `json:"trade_timestamp"`
	OrderTimestamp  int64   `json:"order_timestamp"`
	Timestamp       int64   `json:"timestamp"`
	StreamType      string  `json:"stream_type"`
	Error           error
}

func (s Scraper) upbitMyOrders(callback func(*UpbitMyOrders)) error {
	headers := http.Header{}
	headers.Add("Authorization", fmt.Sprintf("Bearer %s", s.upbit.token))

	conn, _, err := websocket.DefaultDialer.Dial(
		"wss://api.upbit.com/websocket/v1/private",
		headers,
	)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Set up ping/pong handlers
	go keepPing(conn)

	// Create your JSON message
	message := []interface{}{
		map[string]interface{}{
			"ticket": uuid.New().String(), // 예제 ticket
		},
		map[string]interface{}{
			"type": "myOrder",
		},
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(message)
	if err != nil {
		return err
	}

	// Send JSON to WebSocket
	err = conn.WriteMessage(websocket.TextMessage, jsonData)
	if err != nil {
		return err
	}
	// Read messages
	for {
		var order UpbitMyOrders
		_, message, err := conn.ReadMessage()
		if err != nil {
			return err
		}
		err = json.Unmarshal(message, &order)
		if err != nil {
			return err
		}
		if order.State == "done" {
			code, _ := strings.CutPrefix(order.Code, "KRW-")
			order.Code = code
			callback(&order)
		}
	}
	return nil
}

func keepPing(conn *websocket.Conn) {
	const pongWait = 2 * time.Minute
	const pingInterval = (pongWait * 9) / 10 // Send ping before pong timeout

	// Set read deadline
	conn.SetReadDeadline(time.Now().Add(pongWait))

	// Handle pong messages to reset read deadline
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// Start goroutine to send periodic pings
	done := make(chan struct{})
	defer close(done)

	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-done:
			return
		}
	}
}
