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

// todo. 결국 app 패키지가 구현체에 의존하는 구조 개선 필요
// todo. 비지니스 로직을 밖으로 빼는 작업이 필요. 로직이 handler에 가니 불필요하게 객체들이 많이 넘어감
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
