package app

import (
	"fmt"
	"invest/app/handler"
	"invest/app/middleware"
	"invest/db"
	"invest/event"
	"invest/scrape"

	"github.com/gofiber/fiber/v2"
)

func Run(port int, authKey, passKey string, stg *db.Storage, scraper *scrape.Scraper, eh *event.EventHandler) {

	app := fiber.New()

	middleware.SetupMiddleware(app)

	handler.NewAuthHandler(stg, authKey, passKey).InitRoute(app)
	handler.NewAssetHandler(stg, stg, scraper).InitRoute(app)
	handler.NewFundHandler(stg, stg, stg, scraper).InitRoute(app)
	handler.NewInvestHandler(stg, stg, scraper).InitRoute(app)
	handler.NewMarketHandler(stg, stg).InitRoute(app)
	handler.NewCategoryHandler().InitRoute(app)
	handler.NewEventHandler(eh, eh, eh).InitRoute(app)

	app.Get("/shutdown", func(c *fiber.Ctx) error {

		fmt.Println("Shutting Down")
		panic("SHUTDOWN")
	})

	app.Listen(fmt.Sprintf(":%d", port))
}

/*
memo. 커스텀 인코더 지정 가능.
fiber.New(
	fiber.Config(JSONEncoder: customJSONEncoder)
)

func customJSONEncoder(v interface{}) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false) // Disable HTML escaping
	err := encoder.Encode(v)
	return buffer.Bytes(), err
}
*/
