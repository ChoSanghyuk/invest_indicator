package scrape

import (
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

// todo. test 작성
type transmitterMock struct {
}

func (t transmitterMock) Key(target string) string {
	return ""
}

func TestKisApi(t *testing.T) {

	appkey := os.Getenv("appkey")
	appsecret := os.Getenv("appsecret")
	account := os.Getenv("account")
	token := os.Getenv("token")

	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	s, err := NewScraper(
		transmitterMock{},
		WithKIS(&KisConfig{
			AppKey:    appkey,
			AppSecret: appsecret,
			Account:   account,
		}),
		WithKisToken(token),
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Token Generate", func(t *testing.T) {
		token, err := s.kis.KisToken()
		if err != nil {
			t.Error(err)
		}
		t.Log(token)
	})

	t.Run("Stock current Price", func(t *testing.T) {
		stock, err := s.kis.DomesticStockPrice("M04020000") // M04020000 : 금
		if err != nil {
			t.Error(err)
		}
		t.Log(stock.pp, stock.op, stock.hp, stock.lp, stock.ap)
	})

	t.Run("Foreign stock", func(t *testing.T) {
		pp, cp, err := s.kis.ForeignPrice("NAS-MSFT")
		if err != nil {
			t.Error(err)
		}
		t.Log(pp, cp, err)
	})

	t.Run("Nasdaq Index", func(t *testing.T) {
		pp, err := s.kis.Index(Nasdaq)
		if err != nil {
			t.Error(err)
		}
		t.Log(pp)
	})

	t.Run("S&P 500 Index", func(t *testing.T) {
		pp, err := s.kis.Index(Sp500)
		if err != nil {
			t.Error(err)
		}
		t.Log(pp)
	})

	t.Run("Domestic ETF", func(t *testing.T) {
		stock, err := s.kis.DomesticEtfPrice("360750")
		if err != nil {
			t.Error(err)
		}
		t.Log(stock.pp, stock.op, stock.hp, stock.lp, stock.ap)
	})

	t.Run("Foreign ETF", func(t *testing.T) {
		pp, cp, err := s.kis.ForeignPrice("AMS-SPY")
		if err != nil {
			t.Error(err)
		}
		t.Log(pp, cp)
	})

	t.Run("Foreign Period Price", func(t *testing.T) {
		ap, n, err := s.kis.ForeignAvg("NAS-TSLA")
		if err != nil {
			t.Error(err)
		}
		t.Log(ap)
		t.Log(n)
	})

	t.Run("Domestic Daily Ccld", func(t *testing.T) {
		start := time.Now().Add(time.Hour * 24 * -30).Format("20060102") //
		end := time.Now().Add(time.Hour * 24).Format("20060102")
		rtn, err := s.kis.InquireDailyCcld(start, end, "010120")
		if err != nil {
			t.Error(err)
		}
		t.Logf("%s\n", rtn.CtxAreaFk100)
		t.Logf("%s\n", rtn.CtxAreaNk100)
	})

	// t.Run("Domestic stock buy", func(t *testing.T) {
	// 	err := s.kis.DomesticBuy("024950", 1) // 삼천리 자전거
	// 	if err != nil {
	// 		t.Error(err)
	// 	}
	// })

}

func TestKisWs(t *testing.T) {

	appkey := os.Getenv("appkey")
	appsecret := os.Getenv("appsecret")
	htsID := os.Getenv("KIS_HTS_ID")
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	k := NewKis(
		appkey,
		appsecret,
		"",
		htsID,
	)

	// Step 1: Issue WebSocket approval key
	t.Log("Step 1: Issuing WebSocket approval key...")
	approvalResp, err := k.IssueWebSocketApprovalKey()
	if err != nil {
		t.Fatalf("Failed to issue WebSocket approval key: %v", err)
	}
	t.Logf("Approval key received: %s", approvalResp.ApprovalKey[:20]+"...")

	// Step 2: Connect to WebSocket
	t.Log("Step 2: Connecting to WebSocket...")
	err = k.ConnectWebSocket(approvalResp.ApprovalKey)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer k.CloseWebSocket()
	t.Log("WebSocket connected successfully")

	t.Run("realtime Domestic", func(t *testing.T) {
		// Step 3: Subscribe to real-time execution notifications
		t.Log("Step 3: Subscribing to real-time execution notifications...")

		notificationCount := 0
		maxNotifications := 2 // Limit for testing

		err = k.SubscribeRealTimeExecution(func(notification *RealTimeExecutionNotification) {
			notificationCount++
			t.Logf("Notification #%d received:", notificationCount)
			t.Logf("  Order No: %s", notification.OrderNo)
			t.Logf("  Stock Code: %s", notification.StockCode)
			t.Logf("  Sell/Buy: %s (01=Sell, 02=Buy)", notification.SellBuyDiv)
			t.Logf("  Exec Qty: %s", notification.ExecQty)
			t.Logf("  Exec Price: %s", notification.ExecPrice)
			t.Logf("  Exec Time: %s", notification.StockExecTime)
			t.Logf("  Exec YN: %s (1=Order/Revise/Cancel, 2=Execution)", notification.ExecYN)
			t.Logf("  Stock Name: %s", notification.ExecStockName)

			if notificationCount >= maxNotifications {
				t.Logf("Received %d notifications, test complete", notificationCount)
			}
		})

		if err != nil {
			t.Fatalf("Failed to subscribe to real-time execution: %v", err)
		}
	})

	t.Run("realtime Overseas", func(t *testing.T) {
		t.Log("Step 3: Subscribing to overseas real-time execution notifications...")

		notificationCount := 0
		maxNotifications := 5 // Limit for testing

		err = k.SubscribeOverseasRealTimeExecution(htsID, func(notification *OverseasRealTimeExecutionNotification) {
			notificationCount++
			t.Logf("Notification #%d received:", notificationCount)
			t.Logf("  Order No: %s", notification.OrderNo)
			t.Logf("  Stock Code: %s", notification.StockShortCode)
			t.Logf("  Sell/Buy: %s (01=Sell, 02=Buy, 03=Full Sell, 04=Return Buy)", notification.SellBuyDiv)
			t.Logf("  Exec Qty: %s", notification.ExecQty)
			t.Logf("  Exec Price: %s", notification.ExecPrice)
			t.Logf("  Exec Time: %s", notification.StockExecTime)
			t.Logf("  Exec YN: %s (1=Order/Revise/Cancel, 2=Execution)", notification.ExecYN)
			t.Logf("  Stock Name: %s", notification.ExecStockName)
			t.Logf("  Overseas Stock Division: %s", notification.OverseasStockDiv)

			if notificationCount >= maxNotifications {
				t.Logf("Received %d notifications, test complete", notificationCount)
			}
		})

		if err != nil {
			t.Fatalf("Failed to subscribe to overseas real-time execution: %v", err)
		}
	})
	t.Log("Subscription successful")
}
