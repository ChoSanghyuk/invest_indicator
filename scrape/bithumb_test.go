package scrape

import (
	"testing"
)

func TestBithumbApi(t *testing.T) {
	s := Scraper{}
	pp, cp, err := s.bithumbApi("AVAX")
	if err != nil {
		t.Error(err)
	}
	t.Logf("현재가 : %f\n시가: %f", pp, cp)
}
