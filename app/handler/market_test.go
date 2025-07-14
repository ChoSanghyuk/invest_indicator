package handler

import (
	"investindicator/app/middleware"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// router.Get("/", h.Market)
// router.Get("/indicator", h.MarketIndicator)
// router.Post("/", h.ChangeMarketStatus)

func TestMarketHandler(t *testing.T) {

	app := fiber.New()
	middleware.SetupMiddleware(app)

	readerMock := MaketRetrieverMock{}
	writerMock := MarketSaverMock{}
	f := NewMarketHandler(readerMock, writerMock)
	f.InitRoute(app)

	go func() {
		app.Listen(":3000")
	}()

	t.Run("시장단계조회", func(t *testing.T) {
		t.Run("성공테스트-파라미터존재", func(t *testing.T) {
			err := sendReqeust(app, "/market/2024-09-09", "GET", nil, nil)
			assert.NoError(t, err)
		})

		t.Run("성공테스트-파라미터미존재", func(t *testing.T) {
			err := sendReqeust(app, "/market/", "GET", nil, nil)
			assert.NoError(t, err)
		})

		t.Run("실패테스트-잘못된파라미터", func(t *testing.T) {
			err := sendReqeust(app, "/market/202409", "GET", nil, nil)
			assert.Error(t, err)
		})

	})

	t.Run("시장단계저장", func(t *testing.T) {
		t.Run("성공테스트", func(t *testing.T) {
			param := SaveMarketStatusParam{
				Status: 1,
			}
			err := sendReqeust(app, "/market", "POST", param, nil)
			assert.NoError(t, err)
		})

		t.Run("실패테스트-필수파라미터미존재", func(t *testing.T) {
			param := SaveMarketStatusParam{
				// Status: 1,
			}
			err := sendReqeust(app, "/market", "POST", param, nil)
			assert.Error(t, err)
		})

		t.Run("실패테스트-잘못된시장상태값", func(t *testing.T) {
			param := SaveMarketStatusParam{
				Status: 6,
			}
			err := sendReqeust(app, "/market", "POST", param, nil)
			assert.Error(t, err)
		})

	})

	t.Run("시장지표조회", func(t *testing.T) {
		t.Run("성공테스트", func(t *testing.T) {
			err := sendReqeust(app, "/market/indicators/2024-08-29", "GET", nil, nil)
			assert.NoError(t, err)
		})

	})

	app.Shutdown()
}
