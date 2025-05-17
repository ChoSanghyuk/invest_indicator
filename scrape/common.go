package scrape

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

/*
중요!
request body에 nil을 바로 넣는다면, 빈 json 데이터가 들어감.
하지만 nil을 json.Marshal해서 넣는다면, "null"이라는 json 데이터가 형성.
이는 request body에 nil값을 넣는 것과 다른 결과 초래 할 수 있음
*/
func sendRequest(url string, method string, header map[string]string, body map[string]string, response any) error {

	var rb io.Reader
	if body == nil {
		rb = nil
	} else {
		bodyByte, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("error request body marshaling \n%w", err)
		}
		rb = bytes.NewBuffer(bodyByte)
	}

	req, err := http.NewRequest(method, url, rb)
	if err != nil {
		return fmt.Errorf("error making request\n%w", err)
	}

	// Add headers to the request
	for k, v := range header {
		req.Header.Add(k, v)
	}

	// Send the request
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request\n%w", err)
	}

	defer res.Body.Close()

	return json.NewDecoder(res.Body).Decode(response)
}

func (s Scraper) crawl(url string, cssPath string) (string, error) {

	// Send the request
	res, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("error making request\n%w", err)
	}

	defer res.Body.Close()

	// Check the response status
	if res.StatusCode != 200 {
		return "", fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Create a goquery document from the response body
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", fmt.Errorf("error creating document\n%w", err)
	}
	// fmt.Println(doc.Text())

	// fmt.Println(doc.Text())

	var v string
	// Find the elements by the CSS selector
	doc.Find(cssPath).Each(func(i int, s *goquery.Selection) {
		// Extract and print the data
		v = s.Text()
	})

	return v, nil
}
