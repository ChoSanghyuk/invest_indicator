package main

import (
	app "investindicator/app"

	investind "investindicator"
	"investindicator/bot"
	"investindicator/config"
	"investindicator/internal/db"
	"investindicator/scrape"

	"github.com/rs/zerolog"
)

func main() {

	// Create a new instance of the server
	conf, err := config.NewConfig()
	if err != nil {
		panic(err)
	}

	// level, err := conf.LogLevel()
	// if err != nil {
	// 	panic(err)
	// }
	zerolog.SetGlobalLevel(zerolog.InfoLevel) // todo. 글로벌로 설정하면, 다른 모든 logger들에 적용되는지 확인. config로 이동

	ch := make(chan string)

	botConf, err := conf.BotConfig()
	if err != nil {
		panic(err)
	}

	teleBot, err := bot.NewTeleBot(botConf)
	if err != nil {
		panic(err)
	}

	scraper := scrape.NewScraper(conf,
		scrape.WithKIS(conf.KisConfig(teleBot)), // todo. 여기에 봇을 집어넣고, config struct 반환
	)

	db, err := db.NewStorage(conf.Dsn())
	if err != nil {
		panic(err)
	}

	eventHandler := investind.NewInvestIndicator(investind.InvestIndicatorConfig{
		Storage:     db,
		RtPoller:    scraper,
		DailyPoller: scraper,
		Channel:     ch,
	})
	eventHandler.Run()

	go func() {
		app.Run(conf.App.Port, conf.App.JwtKey, conf.App.Passkey, db, scraper, eventHandler)
	}()

	teleBot.Run(ch, conf.App.Port, conf.App.Passkey) // todo. telegram login
}
