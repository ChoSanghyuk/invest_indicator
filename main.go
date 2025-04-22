package main

import (
	"invest/app"

	"invest/bot"
	"invest/config"
	"invest/db"
	"invest/event"
	"invest/scrape"

	"github.com/rs/zerolog"
)

func main() {

	// Create a new instance of the server
	conf, err := config.NewConfig()
	if err != nil {
		panic(err)
	}

	level, err := conf.LogLevel()
	if err != nil {
		panic(err)
	}
	zerolog.SetGlobalLevel(level) // todo. 글로벌로 설정하면, 다른 모든 logger들에 적용되는지 확인

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

	eventHandler := event.NewEventHandler(event.EventHandlerConfig{
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
