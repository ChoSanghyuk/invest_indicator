package scrape

import (
	"testing"
)

func TestUpbitApi(t *testing.T) {
	s := Scraper{}
	pp, cp, err := s.upbitApi("KRW-AVAX")
	if err != nil {
		t.Error(err)
	}
	t.Logf("현재가 : %f\n시가: %f", pp, cp)
}
