package config

import (
	_ "embed"
	"fmt"

	"investindicator/bot"
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

func (c Config) InitKIS(key string) (err error) {
	*c.KIS["appkey"], err = util.Decrypt([]byte(key), *c.KIS["appkey"])
	if err != nil {
		return err
	}

	*c.KIS["appsecret"], err = util.Decrypt([]byte(key), *c.KIS["appsecret"])
	if err != nil {
		return err
	}
	return nil
}

// 복호화 키 전달
type KeyPasser interface {
	InitKey(err error) string
}

func (c Config) KisConfig(keyPasser KeyPasser) *scrape.KisConfig {

	var appKey string
	var appSecret string
	var err error

	for appKey == "" || appSecret == "" || err != nil {
		key := keyPasser.InitKey(err)
		appKey, err = util.Decrypt([]byte(key), *c.KIS["appkey"])
		if err != nil {
			continue
		}

		appSecret, err = util.Decrypt([]byte(key), *c.KIS["appsecret"])
	}

	return &scrape.KisConfig{
		AppKey:    appKey,
		AppSecret: appSecret,
	}
}

func (c Config) Dsn() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", c.Db.User, c.Db.Password, c.Db.IP, c.Db.Port, c.Db.Scheme)
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
