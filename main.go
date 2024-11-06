package main

import (
	"invest/app"

	"invest/bot"
	"invest/config"
	"invest/db"
	"invest/event"
	"invest/scrape"
	"strconv"

	"log"

	"github.com/robfig/cron"
)

const (
	AssetSpec  = "0 */15 8-23 * * 1-5"
	CoinSpec   = "0 */15 8-23 * * 0,6"
	EstateSpec = "0 */15 9-17 * * 1-5"
	IndexSpec  = "0 3 9 * * 1-5" // todo. 9시 3분이랑 8시 3분이랑 값이 같은지 확인
	EmaSpec    = "0 3 9 * * 2-6" // 화~토
)

func main() {
	// Create a new instance of the server

	conf, err := config.NewConfig()
	if err != nil {
		panic(err)
	}

	ch := make(chan string)

	chatId, err := strconv.ParseInt(conf.Telegram.ChatId, 10, 64)
	if err != nil {
		panic(err)
	}
	teleBot, err := bot.NewTeleBot(conf.Telegram.Token, chatId)
	if err != nil {
		panic(err)
	}

	for {
		key := teleBot.InitKey()
		err = conf.InitKIS(key)

		if err != nil {
			teleBot.SendMessage(err.Error())
		} else {
			break
		}
	}

	go func() {
		teleBot.Listen(ch)
	}()

	scraper := scrape.NewScraper(conf,
		scrape.WithKIS(conf.KisAppKey(), conf.KisAppSecret()),
	)

	db, err := db.NewStorage(conf.Dsn())
	if err != nil {
		panic(err)
	}
	event := event.NewEvent(db, scraper, scraper)

	c := cron.New()
	c.AddFunc(AssetSpec, func() { event.AssetEvent(ch) })
	c.AddFunc(CoinSpec, func() { event.CoinEvent(ch) })
	c.AddFunc(EstateSpec, func() { event.RealEstateEvent(ch) })
	c.AddFunc(IndexSpec, func() { event.IndexEvent(ch) })
	c.AddFunc(EmaSpec, func() { event.EmaUpdateEvent(ch) })
	c.Start()

	go func() {
		app.Run(db, scraper)
	}()

	for true {
		msg := <-ch
		teleBot.SendMessage(msg)
		log.Println(msg)
	}
}
