package handler

import (
	"invest/app/middleware"
	m "invest/model"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestInvestGetHandler(t *testing.T) {

	const initAmount float64 = 100000
	const dollarPrice float64 = 1400

	app := fiber.New()
	middleware.SetupMiddleware(app)

	readerMock := NewADefaultssetRetrieverMock()
	writerMock := NewInvestSaverMock(initAmount)
	exMock := NewExchageRateGetterMock(dollarPrice)

	f := NewInvestHandler(readerMock, writerMock, exMock)
	f.InitRoute(app)
	go func() {
		app.Listen(":3000")
	}()

	t.Run("투자 이력 저장", func(t *testing.T) {
		writerMock.reset()

		param := SaveInvestParam{
			FundId:  1,
			AssetId: 3,
			Price:   10000,
			Count:   3,
		}
		err := sendReqeust(app, "/invest", "POST", param, nil)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(writerMock.invests))
		assert.Equal(t, 2, len(writerMock.summaries))
		assert.Equal(t, initAmount-10000*3, writerMock.summaries[0].Sum)
		writerMock.prettyPrint()
	})

	t.Run("달러 환전", func(t *testing.T) {
		writerMock.reset()
		param := SaveInvestParam{
			FundId:  1,
			AssetId: 2,
			Price:   1500,
			Count:   10,
		}
		err := sendReqeust(app, "/invest", "POST", param, nil)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(writerMock.invests))
		assert.Equal(t, 2, len(writerMock.summaries))
		assert.Equal(t, initAmount-1500*10, writerMock.summaries[0].Sum)
		writerMock.prettyPrint()
	})

	t.Run("달러 자산 구매", func(t *testing.T) {
		writerMock.reset()
		writerMock.summaries = append(writerMock.summaries, m.InvestSummary{
			ID:      1,
			FundID:  1,
			AssetID: 2,
			Sum:     100 * dollarPrice,
			Count:   100,
		})
		param := SaveInvestParam{
			FundId:  1,
			AssetId: 4,
			Price:   1,
			Count:   10,
		}
		err := sendReqeust(app, "/invest", "POST", param, nil)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(writerMock.invests))
		assert.Equal(t, 3, len(writerMock.summaries))
		assert.Equal(t, 100*dollarPrice-dollarPrice*10, writerMock.summaries[1].Sum)
		writerMock.prettyPrint()
	})

	app.Shutdown()
}
