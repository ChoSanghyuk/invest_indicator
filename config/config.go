package config

import (
	_ "embed"
	"fmt"
	"invest/util"

	"gopkg.in/yaml.v3"
)

//go:embed config.yaml
var configByte []byte

type Config struct {
	Api      map[string]apiConfig   `yaml:"api"`
	Crawl    map[string]crawlConfig `yaml:"crawl"`
	Telegram struct {
		ChatId string `yaml:"chatId"`
		Token  string `yaml:"token"`
	} `yaml:"telegram"`
	Key struct {
		KIS map[string]*string `yaml:"KIS"`
	} `yaml:"key"`

	Db struct { // "root:root@tcp(127.0.0.1:3300)/investdb?charset=utf8mb4&parseTime=True&loc=Local"
		User     string `yaml:"user"`
		Password string `yaml:"pwd"`
		IP       string `yaml:"ip"`
		Port     string `yaml:"port"`
		Scheme   string `yaml:"scheme"`
	} `yaml:"db"`
}

type apiConfig struct {
	Url    string            `yaml:"url"`
	Header map[string]string `yaml:"header"`
}

type crawlConfig struct {
	Url     string `yaml:"url"`
	CssPath string `yaml:"css-path"`
}

func NewConfig() (*Config, error) {

	var ConfigInfo Config = Config{}

	err := yaml.Unmarshal(configByte, &ConfigInfo)
	if err != nil {
		return nil, err
	}

	// util.Decode(&ConfigInfo.Gold.API.ApiKey)
	util.Decode(&ConfigInfo.Telegram.ChatId)
	util.Decode(&ConfigInfo.Telegram.Token)
	// util.Decode(ConfigInfo.Key.KIS["appkey"])
	// util.Decode(ConfigInfo.Key.KIS["appsecret"])

	return &ConfigInfo, nil
}

func (c Config) InitKIS(key string) (err error) {
	*c.Key.KIS["appkey"], err = decrypt([]byte(key), *c.Key.KIS["appkey"])
	if err != nil {
		return err
	}

	*c.Key.KIS["appsecret"], err = decrypt([]byte(key), *c.Key.KIS["appsecret"])
	if err != nil {
		return err
	}
	return nil
}

func (c Config) KisAppKey() string {
	return *c.Key.KIS["appkey"]
}

func (c Config) KisAppSecret() string {
	return *c.Key.KIS["appsecret"]
}

func (c Config) Dsn() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", c.Db.User, c.Db.Password, c.Db.IP, c.Db.Port, c.Db.Scheme)
}

func (c Config) ApiBaseUrl(target string) string {
	return c.Api[target].Url
}

func (c Config) ApiHeader(target string) map[string]string {
	return c.Api[target].Header
}

func (c Config) CrawlUrlCasspath(target string) (url string, cssPath string) {
	return c.Crawl[target].Url, c.Crawl[target].CssPath
}
