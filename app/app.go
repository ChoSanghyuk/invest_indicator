package app

import (
	"fmt"
	"invest/app/handler"
	"invest/db"
	"invest/scrape"

	"github.com/gofiber/fiber/v2"
)

func Run(stg *db.Storage, scraper *scrape.Scraper) {

	app := fiber.New()

	handler.NewAssetHandler(stg, stg, scraper).InitRoute(app)
	handler.NewFundHandler(stg, stg, scraper).InitRoute(app)
	handler.NewInvestHandler(stg, stg, scraper).InitRoute(app)
	handler.NewMarketHandler(stg, stg).InitRoute(app)

	app.Get("/shutdown", func(c *fiber.Ctx) error {

		fmt.Println("Shutting Down")
		panic("SHUTDOWN")
		// go func() {
		// 	if err := app.Shutdown(); err != nil {
		// 		panic("SHUTDOWN ERROR")
		// 	}
		// }()
		// return c.SendString("Shutting Down")
	})

	app.Listen(":3000")

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
