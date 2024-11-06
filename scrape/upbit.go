package scrape

import (
	"errors"
	"fmt"
	"net/http"
)

func (s Scraper) upbitApi(sym string) (float64, float64, error) {

	url := s.t.ApiBaseUrl("upbit")
	if url == "" {
		return 0, 0, errors.New("URL 미존재")
	}
	url = fmt.Sprintf(url, sym)

	var rtn []map[string]any
	err := sendRequest(url, http.MethodGet, nil, nil, &rtn)
	if err != nil {
		return 0, 0, err
	}

	return rtn[0]["trade_price"].(float64), rtn[0]["opening_price"].(float64), nil // 시가 = 전날 종가
}
