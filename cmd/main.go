package main

import (
	investind "investindicator"
	app "investindicator/app"
	"investindicator/blockchain"
	blackhole "investindicator/blockchain/blackhole"
	"investindicator/blockchain/pkg/txlistener"
	"investindicator/blockchain/uniswap"
	"investindicator/bot"
	"investindicator/config"
	"investindicator/internal/db"
	"investindicator/scrape"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
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

	botConfs, err := conf.BotConfigs()
	if err != nil {
		panic(err)
	}

	teleBotGroup := bot.NewTeleBotGroup(botConfs)

	scraper, err := scrape.NewScraper(conf,
		scrape.WithKIS(conf.KisConfig(teleBotGroup.Bot(0))), // todo. 여기에 봇을 집어넣고, config struct 반환
		scrape.WithUpbitToken(conf.UpbitConfig(teleBotGroup.Bot(0))),
	)
	if err != nil {
		panic(err)
	}

	db, err := db.NewStorage(conf.MysqlConfig(), conf.RedisConfig())
	if err != nil {
		panic(err)
	}

	/* blockchain */
	client, err := ethclient.Dial(conf.Blockchain.Blackhole.Url)
	if err != nil {
		panic(err)
	}
	listener := txlistener.NewTxListener(
		client,
		txlistener.WithPollInterval(2*time.Second),
		txlistener.WithTimeout(5*time.Minute),
	)

	us, err := uniswap.NewUniswapClient(conf.UniswapConfig(teleBotGroup.Bot(0)))
	if err != nil {
		panic(err)
	}

	bd, err := blackhole.NewBlackhole(
		client,
		conf.BlackholeConfig(teleBotGroup.Bot(0)),
		listener,
		db,
	)
	if err != nil {
		panic(err)
	}

	bt := blockchain.NewBlockChainTrader(us, bd, conf.ToStrategyConfig())

	eventHandler := investind.NewInvestIndicator(db, scraper, scraper, bt, teleBotGroup)
	eventHandler.Run()

	teleBotGroup.RunAll(conf.App.Port, conf.App.Passkey) // todo. telegram login

	app.Run(conf.App.Port, conf.App.JwtKey, conf.App.Passkey, db, scraper, eventHandler)
}
