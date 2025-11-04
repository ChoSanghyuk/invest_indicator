package scrape

import (
	"fmt"
	"net/http"
)

const bithumbUrlForm = "https://api.bithumb.com/v1/candles/days?market=%s&count=1"

func (s Scraper) bithumbApi(sym string) (float64, float64, error) {

	url := fmt.Sprintf(bithumbUrlForm, sym)

	var rtn []map[string]any
	err := sendRequest(url, http.MethodGet, nil, nil, &rtn)
	if err != nil {
		return 0, 0, err
	}

	return rtn[0]["trade_price"].(float64), rtn[0]["opening_price"].(float64), nil // 시가 = 전날 종가
}
