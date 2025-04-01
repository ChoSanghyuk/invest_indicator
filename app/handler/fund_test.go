package handler

import (
	"fmt"
	"invest/app/middleware"
	m "invest/model"
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
	f := NewFundHandler(readerMock, writerMock, investMock, exGetterMock)
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

	app.Shutdown()
}
