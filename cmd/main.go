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

	level, err := conf.LogLevel()
	if err != nil {
		panic(err)
	}
	/*
		memo.
		zerolog.SetGlobalLevel()는 이후에 생성되는 모든 zerolog.Logger의 로그 레벨을 설정함.
		단, 이 프로그램의 경우, mysql.go에서는 별도의 gorm logger을 사용하기 때문에 영향을 받지 않음.
	*/
	zerolog.SetGlobalLevel(level) // todo. 글로벌로 설정하면, 다른 모든 logger들에 적용되는지 확인. config로 이동

	ch := make(chan string)

	botConf, err := conf.BotConfig()
	if err != nil {
		panic(err)
	}

	teleBot, err := bot.NewTeleBot(botConf)
	if err != nil {
		panic(err)
	}

	scraper, err := scrape.NewScraper(conf,
		scrape.WithKIS(conf.KisConfig(teleBot)), // todo. 여기에 봇을 집어넣고, config struct 반환
	)
	if err != nil {
		panic(err)
	}

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
