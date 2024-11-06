package config

import (
	"testing"
)

func TestConfigNew(t *testing.T) {

	conf, _ := NewConfig()
	t.Log(conf.CrawlUrlCasspath("exchangeRate"))
	// t.Log(*conf.Key.KIS["appsecret"])

}
