package handler

import (
	"invest/app/middleware"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestInvestGetHandler(t *testing.T) {

	app := fiber.New()
	middleware.SetupMiddleware(app)

	readerMock := AssetRetrieverMock{}
	writerMock := InvestSaverMock{}
	exMock := ExchageRateGetterMock{}
	f := NewInvestHandler(readerMock, writerMock, exMock)
	f.InitRoute(app)
	go func() {
		app.Listen(":3000")
	}()

	t.Run("투자 이력 저장", func(t *testing.T) {
		t.Run("성공 테스트", func(t *testing.T) {
			param := SaveInvestParam{
				FundId:  1,
				AssetId: 1,
				Price:   56532,
				Count:   3,
			}
			err := sendReqeust(app, "/invest", "POST", param, nil)
			assert.NoError(t, err)
		})

	})

	app.Shutdown()
}
