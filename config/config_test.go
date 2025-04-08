package config

import "testing"

func TestConfigInit(t *testing.T) {
	conf, err := NewConfig()
	if err != nil {
		t.Error(err)
	}

	t.Logf("%+v", conf)
}
