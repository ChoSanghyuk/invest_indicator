package config

import (
	_ "embed"
	"math/big"

	"investindicator/blockchain"
	"investindicator/bot"
	"investindicator/internal/db"
	"investindicator/internal/util"
	"investindicator/scrape"
	"strconv"

	"github.com/rs/zerolog"
	"gopkg.in/yaml.v3"
)

//go:embed config.yaml
var configByte []byte

type Config struct {
	Log string `yaml:"log"`
	App struct {
		Port    int    `yaml:"port"`
		JwtKey  string `yaml:"jwtkey"`
		Passkey string `yaml:"passkey"`
	} `yaml:"app"`
	ApiKey   map[string]string `yaml:"api-key"`
	Telegram struct {
		ChatId string `yaml:"chatId"`
		Token  string `yaml:"token"`
	} `yaml:"telegram"`

	KIS map[string]*string `yaml:"KIS"`

	Db struct {
		User     string `yaml:"user"`
		Password string `yaml:"pwd"`
		IP       string `yaml:"ip"`
		Port     string `yaml:"port"`
		Scheme   string `yaml:"scheme"`
	} `yaml:"db"`
	Blockchain struct {
		Uniswap struct {
			Url             string `yaml:"url"`
			Pk              string `yaml:"pk"`
			UniversalRouter string `yaml:"universalrouter"`
			Permit          string `yaml:"permit"`
			GasLimit        int    `yaml:"gaslimit"`
		} `yaml:"uniswap"`
	} `yaml:"blockchain"`
	decryptKey string
}

func NewConfig() (*Config, error) {

	var ConfigInfo Config = Config{}

	err := yaml.Unmarshal(configByte, &ConfigInfo)
	if err != nil {
		return nil, err
	}

	decode(&ConfigInfo)

	return &ConfigInfo, nil
}

func (c Config) LogLevel() (zerolog.Level, error) {

	level, err := zerolog.ParseLevel(c.Log)
	if err != nil {
		return zerolog.InfoLevel, err // Default로는 Info 레벨 설정
	}

	return level, nil
}

func (c Config) BotConfig() (*bot.TeleBotConfig, error) {

	chatId, err := strconv.ParseInt(c.Telegram.ChatId, 10, 64)
	if err != nil {
		return nil, err
	}

	return &bot.TeleBotConfig{
		Token:  c.Telegram.Token,
		ChatId: chatId,
	}, nil
}

func (c Config) UniswapConfig(keyPasser KeyPasser) *blockchain.UniswapClientConfig {

	var pk string
	var err error
	var key string

init:
	if c.decryptKey == "" { // 키 등록이 안 된 경우에는 키 입력 받기
		key = keyPasser.InitKey(err)
	}
	pk, err = util.Decrypt([]byte(key), c.Blockchain.Uniswap.Pk)
	if err != nil { // 오류인 경우, 키 입력 반복
		goto init
	}
	c.decryptKey = key // 오류 없이 통과한 경우에만 등록

	return blockchain.NewUniswapClientConfig(
		c.Blockchain.Uniswap.Url,
		pk,
		c.Blockchain.Uniswap.UniversalRouter,
		c.Blockchain.Uniswap.Permit,
		big.NewInt(int64(c.Blockchain.Uniswap.GasLimit)),
	)
}

// func (c Config) InitKIS(key string) (err error) {
// 	*c.KIS["appkey"], err = util.Decrypt([]byte(key), *c.KIS["appkey"])
// 	if err != nil {
// 		return err
// 	}
// 	*c.KIS["appsecret"], err = util.Decrypt([]byte(key), *c.KIS["appsecret"])
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// 복호화 키 전달
type KeyPasser interface {
	InitKey(err error) string
}

func (c *Config) KisConfig(keyPasser KeyPasser) *scrape.KisConfig {

	var err error
	var key string = c.decryptKey

init:
	if c.decryptKey == "" {
		key = keyPasser.InitKey(err)
	}
	appKey, err := util.Decrypt([]byte(key), *c.KIS["appkey"])
	if err != nil {
		goto init
	}
	appSecret, err := util.Decrypt([]byte(key), *c.KIS["appsecret"])
	if err != nil {
		goto init
	}

	c.decryptKey = key
	return &scrape.KisConfig{
		AppKey:    appKey,
		AppSecret: appSecret,
		Account:   *c.KIS["account"],
	}
}

func (c Config) StgConfig() *db.StgConfig {
	return db.NewStgConfig(c.Db.User, c.Db.Password, c.Db.IP, c.Db.Port, c.Db.Scheme)
}

func (c Config) Key(target string) string {
	return c.ApiKey[target]
}

func decode(conf *Config) {
	util.Decode(&conf.Telegram.ChatId)
	util.Decode(&conf.Telegram.Token)
	util.Decode(&conf.App.JwtKey)
	util.Decode(&conf.App.Passkey)
}
