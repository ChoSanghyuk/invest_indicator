package scrape

import (
	"fmt"
	"net/http"
)

const upbitUrlForm = "https://api.upbit.com/v1/candles/days?market=%s&count=1"

func (s Scraper) upbitApi(sym string) (float64, float64, error) {

	url := fmt.Sprintf(upbitUrlForm, "KRW-"+sym)

	var rtn []map[string]any
	err := sendRequest(url, http.MethodGet, nil, nil, &rtn)
	if err != nil {
		return 0, 0, err
	}

	return rtn[0]["trade_price"].(float64), rtn[0]["opening_price"].(float64), nil // 시가 = 전날 종가
}
