package handler

import (
	"fmt"
	"investindicator/app/middleware"
	m "investindicator/internal/model"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestFundHandler(t *testing.T) {

	app := fiber.New()
	middleware.SetupMiddleware(app)

	readerMock := &FundRetrieverMock{}
	writerMock := &FundWriterMock{}
	investMock := &InvestRetrieverMock{}
	exGetterMock := &ExchageRateGetterMock{}
	marketMock := &MaketRetrieverMock{}
	f := NewFundHandler(readerMock, writerMock, investMock, exGetterMock, marketMock)
	f.InitRoute(app)

	go func() {
		app.Listen(":3000")
	}()

	t.Run("전체 자금별 총액", func(t *testing.T) {
		t.Run("성공 테스트", func(t *testing.T) {
			readerMock.isli = []m.InvestSummary{
				{ID: 1, FundID: 1, Fund: m.Fund{Name: "공용자금"}, AssetID: 1, Sum: 10000},
				{ID: 2, FundID: 1, Fund: m.Fund{Name: "공용자금"}, AssetID: 2, Sum: 12000},
				{ID: 3, FundID: 1, Fund: m.Fund{Name: "공용자금"}, AssetID: 3, Sum: 15000},
				{ID: 4, FundID: 2, Fund: m.Fund{Name: "퇴직연금"}, AssetID: 1, Sum: 10000},
				{ID: 5, FundID: 2, Fund: m.Fund{Name: "퇴직연금"}, AssetID: 2, Sum: 20000},
			}

			resp := make(map[uint]TotalStatusResp)
			err := sendReqeust(app, "/funds", "GET", nil, &resp)

			assert.NoError(t, err)
			if resp[1].Amount != 37000 {
				t.Error()
			}
			if resp[2].Amount != 30000 {
				t.Error()
			}
			t.Logf("%+v\n", resp)
		})
	})

	t.Run("자금 추가", func(t *testing.T) {
		t.Run("성공 테스트", func(t *testing.T) {
			param := AddFundReq{
				Name: "신규 자금",
			}
			err := sendReqeust(app, "/funds", "POST", param, nil)
			assert.NoError(t, err)
		})

	})

	t.Run("자금 투자 이력 조회", func(t *testing.T) {
		t.Run("성공 테스트", func(t *testing.T) {
			readerMock.il = []m.Invest{
				{ID: 1, FundID: 1, AssetID: 1, Price: 7800, Count: 5},
				{ID: 2, FundID: 2, AssetID: 1, Price: 7800, Count: 3},
			}

			var resp []m.Invest
			err := sendReqeust(app, "/funds/1/hist", "GET", nil, &resp)
			assert.NoError(t, err)

			for _, iv := range resp {
				if iv.FundID != 1 {
					t.Error()
				}
			}
			t.Logf("\n%+v\n", resp)
		})

	})

	t.Run("이력_조회_날짜", func(t *testing.T) {
		t.Run("성공 테스트", func(t *testing.T) {

			readerMock.il = []m.Invest{
				{ID: 1, FundID: 1, AssetID: 1, Price: 7800, Count: 5, Model: gorm.Model{CreatedAt: time.Now()}},
				{ID: 2, FundID: 2, AssetID: 1, Price: 7800, Count: 3, Model: gorm.Model{CreatedAt: time.Now().AddDate(0, -1, 0)}},
			}

			var resp []HistResponse
			err := sendReqeust(app,
				fmt.Sprintf("/funds/1/hist?start=%s&end=%s",
					time.Now().AddDate(0, 0, -1),
					time.Now().AddDate(0, 0, 1),
				),
				"GET",
				nil,
				&resp)
			assert.NoError(t, err)

			for _, iv := range resp {
				if iv.FundId != 1 {
					t.Error()
				}
			}
			t.Logf("\n%+v\n", resp)
		})

	})

	t.Run("자금별 투자 종목 총액 조회", func(t *testing.T) {
		t.Run("성공 테스트", func(t *testing.T) {

			readerMock.isli = []m.InvestSummary{
				{ID: 1, FundID: 1, Fund: m.Fund{Name: "공용자금"}, AssetID: 1, Sum: 10000},
				{ID: 2, FundID: 1, Fund: m.Fund{Name: "공용자금"}, AssetID: 2, Sum: 12000},
				{ID: 3, FundID: 1, Fund: m.Fund{Name: "공용자금"}, AssetID: 3, Sum: 15000},
				{ID: 4, FundID: 2, Fund: m.Fund{Name: "퇴직연금"}, AssetID: 1, Sum: 10000},
				{ID: 5, FundID: 2, Fund: m.Fund{Name: "퇴직연금"}, AssetID: 2, Sum: 20000},
			}

			var resp []m.InvestSummary

			err := sendReqeust(app, "/funds/1/assets", "GET", nil, &resp)
			assert.NoError(t, err)
			if len(readerMock.isli) != len(resp) {
				t.Error()
			}
			t.Logf("\n%+v\n", resp)
		})
	})

	t.Run("자금 비중 조회", func(t *testing.T) {
		t.Run("성공 테스트", func(t *testing.T) {
			readerMock.isli = []m.InvestSummary{
				{ID: 1, FundID: 1, Fund: m.Fund{Name: "공용자금"}, Asset: m.Asset{Category: m.Dollar}, AssetID: 1, Count: 10, Sum: 10000},
				{ID: 2, FundID: 1, Fund: m.Fund{Name: "공용자금"}, Asset: m.Asset{Category: m.Dollar}, AssetID: 2, Count: 10, Sum: 12000},
				{ID: 3, FundID: 1, Fund: m.Fund{Name: "공용자금"}, Asset: m.Asset{Category: m.Leverage}, AssetID: 3, Count: 10, Sum: 15000},
				// {ID: 4, FundID: 2, Fund: m.Fund{Name: "퇴직연금"}, Asset: m.Asset{Category: m.Dollar}, AssetID: 1, Count: 10, Sum: 10000},
				// {ID: 5, FundID: 2, Fund: m.Fund{Name: "퇴직연금"}, Asset: m.Asset{Category: m.Leverage}, AssetID: 2, Count: 10, Sum: 20000},
			}

			var resp fundPortionResponse

			err := sendReqeust(app, "/funds/1/portion", "GET", nil, &resp)
			assert.NoError(t, err)
			t.Logf("\n%+v\n", resp)
		})
	})

	t.Run("수익률 조회", func(t *testing.T) {
		readerMock.isli = []m.InvestSummary{
			{ID: 1, FundID: 1, AssetID: 3, Count: 0.00354433, Sum: 577520.21886},
		}

		investMock.invests = []m.Invest{
			{ID: 1, FundID: 1, AssetID: 3, Price: 150511000, Count: -0.00362779},
			{ID: 2, FundID: 1, AssetID: 3, Price: 150511000, Count: 0.00362779},
			{ID: 3, FundID: 1, AssetID: 3, Price: 141070000, Count: 0.00354433},
		}

		rtn := f.profitRateOfAsset(&readerMock.isli[0])
		t.Logf("rtn : %s", rtn)
	})

	t.Run("AvailableAmounts 테스트", func(t *testing.T) {
		t.Run("MAJOR_BEAR 시장에서 가용 금액 계산", func(t *testing.T) {
			// Setup: MAJOR_BEAR 시장 (MaxVolatileAssetRate = 0.15)
			*marketMock = MaketRetrieverMock{err: nil, marketStatus: 1} // MAJOR_BEAR = 1

			// 환율 1300원
			*exGetterMock = ExchageRateGetterMock{exchangeRate: 1300.0}

			// 자금 구성: 안전자산(원화) 50,000원 + 변동자산(USD) 100달러 = 총 180,000원
			// 변동자산 비율: 130,000 / 180,000 = 0.722... (현재 비율이 높음)
			readerMock.isli = []m.InvestSummary{
				{ID: 1, FundID: 1, Asset: m.Asset{Category: m.Won, Currency: m.KRW.String()}, Count: 50000, Sum: 50000},       // 안전자산
				{ID: 2, FundID: 1, Asset: m.Asset{Category: m.DomesticStock, Currency: m.USD.String()}, Count: 100, Sum: 100}, // 변동자산
			}

			var resp float64
			err := sendReqeust(app, "/funds/1/available_amounts", "GET", nil, &resp)

			assert.NoError(t, err)
			// 기대값: 0.15 * 180000 - 130000 = 27000 - 130000 = -103000 (음수는 매도 필요)
			expectedAmount := 0.15*180000 - 130000
			assert.Equal(t, expectedAmount, resp)
			t.Logf("Available amount: %.2f (negative means need to sell)", resp)
		})

		t.Run("BULL 시장에서 가용 금액 계산", func(t *testing.T) {
			// Setup: BULL 시장 (MaxVolatileAssetRate = 0.3)
			*marketMock = MaketRetrieverMock{err: nil, marketStatus: 4} // BULL = 4

			// 환율 1300원
			*exGetterMock = ExchageRateGetterMock{exchangeRate: 1300.0}

			// 자금 구성: 안전자산 100,000원 + 변동자산 50달러 = 총 165,000원
			// 변동자산 비율: 65,000 / 165,000 = 0.394... (약간 높음, 하지만 BULL 시장이므로 괜찮을 수 있음)
			readerMock.isli = []m.InvestSummary{
				{ID: 1, FundID: 1, Asset: m.Asset{Category: m.Dollar, Currency: m.KRW.String()}, Count: 100000, Sum: 100000}, // 안전자산
				{ID: 2, FundID: 1, Asset: m.Asset{Category: m.DomesticStock, Currency: m.USD.String()}, Count: 50, Sum: 50},  // 변동자산
			}

			var resp float64
			err := sendReqeust(app, "/funds/1/available_amounts", "GET", nil, &resp)

			assert.NoError(t, err)
			// 기대값: 0.3 * 165000 - 65000 = 49500 - 65000 = -15500 (여전히 매도 필요)
			expectedAmount := 0.3*165000 - 65000
			assert.InDelta(t, expectedAmount, resp, 0.01) // Allow small floating point differences
			t.Logf("Available amount: %.2f", resp)
		})

		t.Run("순수 안전자산만 보유한 경우", func(t *testing.T) {
			*marketMock = MaketRetrieverMock{err: nil, marketStatus: 1} // MAJOR_BEAR = 1
			*exGetterMock = ExchageRateGetterMock{exchangeRate: 1300.0}

			// 모든 자산이 안전자산 (원화 + 달러)
			readerMock.isli = []m.InvestSummary{
				{ID: 1, FundID: 1, Asset: m.Asset{Category: m.Won, Currency: m.KRW.String()}, Count: 100000, Sum: 100000},
				{ID: 2, FundID: 1, Asset: m.Asset{Category: m.Dollar, Currency: m.USD.String()}, Count: 100, Sum: 100}, // USD → 130,000원
			}

			var resp float64
			err := sendReqeust(app, "/funds/1/available_amounts", "GET", nil, &resp)

			assert.NoError(t, err)
			// 기대값: 0.15 * 230000 - 0 = 34500 (매수 가능)
			expectedAmount := 0.15*230000 - 0
			assert.Equal(t, expectedAmount, resp)
			t.Logf("Available amount: %.2f (positive means can buy)", resp)
		})

		t.Run("Count가 0인 자산 제외 테스트", func(t *testing.T) {
			*marketMock = MaketRetrieverMock{err: nil, marketStatus: 1} // MAJOR_BEAR = 1
			*exGetterMock = ExchageRateGetterMock{exchangeRate: 1300.0}

			// Count가 0인 자산은 계산에서 제외되어야 함
			readerMock.isli = []m.InvestSummary{
				{ID: 1, FundID: 1, Asset: m.Asset{Category: m.Won, Currency: m.KRW.String()}, Count: 0, Sum: 50000},               // Count=0, 제외
				{ID: 2, FundID: 1, Asset: m.Asset{Category: m.Dollar, Currency: m.KRW.String()}, Count: 100000, Sum: 100000},      // 안전자산
				{ID: 3, FundID: 1, Asset: m.Asset{Category: m.DomesticStock, Currency: m.KRW.String()}, Count: 50000, Sum: 50000}, // 변동자산
			}

			var resp float64
			err := sendReqeust(app, "/funds/1/available_amounts", "GET", nil, &resp)

			assert.NoError(t, err)
			// Count=0인 첫 번째 자산은 제외, 총액 150000, 변동자산 50000
			// 기대값: 0.15 * 150000 - 50000 = 22500 - 50000 = -27500
			expectedAmount := 0.15*150000 - 50000
			assert.Equal(t, expectedAmount, resp)
			t.Logf("Available amount: %.2f", resp)
		})

		t.Run("파라미터 오류 테스트", func(t *testing.T) {
			var resp interface{}
			err := sendReqeust(app, "/funds/invalid/available_amounts", "GET", nil, &resp)

			assert.Error(t, err)
			t.Logf("Expected error for invalid parameter: %v", err)
		})

		t.Run("데이터베이스 오류 테스트", func(t *testing.T) {
			// Fund 조회 오류 시뮬레이션
			readerMock.err = fmt.Errorf("database connection error")

			var resp interface{}
			err := sendReqeust(app, "/funds/1/available_amounts", "GET", nil, &resp)

			assert.Error(t, err)
			readerMock.err = nil // Reset for other tests
			t.Logf("Expected error for database error: %v", err)
		})

		t.Run("마켓 상태 조회 오류 테스트", func(t *testing.T) {
			*marketMock = MaketRetrieverMock{err: fmt.Errorf("market status error"), marketStatus: 0}

			readerMock.isli = []m.InvestSummary{
				{ID: 1, FundID: 1, Asset: m.Asset{Category: m.Won, Currency: m.KRW.String()}, Count: 100000, Sum: 100000},
			}

			var resp interface{}
			err := sendReqeust(app, "/funds/1/available_amounts", "GET", nil, &resp)

			assert.Error(t, err)
			*marketMock = MaketRetrieverMock{err: nil, marketStatus: 0} // Reset for other tests
			t.Logf("Expected error for market status error: %v", err)
		})
	})

	app.Shutdown()
}
