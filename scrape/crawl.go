package scrape

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
)

func crawl(url string, cssPath string) (string, error) {

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

func crawlWithHeader(url string, cssPath string) (string, error) {

	// Send the request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "ko-KR,ko;q=0.8,en-US;q=0.5,en;q=0.3")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error requesting\n%w", err)
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

func crawlSpaBody(url string) (*goquery.Document, error) {

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var htmlContent string
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.Sleep(3*time.Second),
		chromedp.OuterHTML("html", &htmlContent),
	)
	if err != nil {
		return nil, nil
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("error creating document\n%w", err)
	} else if doc == nil {
		return nil, fmt.Errorf("no document. check chrome browser exists\n%w", err)
	}

	return doc, nil
}

func crawlSpaBodyAvoidingClaudFlare(url string) (*goquery.Document, error) {

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
		chromedp.WindowSize(1920, 1080),
	)

	ctx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	var htmlContent string
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Sleep(3134*time.Millisecond), // Wait for Cloudflare to complete
		chromedp.ActionFunc(func(ctx context.Context) error { // Check if we're still on Cloudflare page
			var title string
			chromedp.Title(&title).Do(ctx)
			if strings.Contains(title, "Cloudflare") {
				fmt.Println("Still on Cloudflare page, waiting longer...")
				time.Sleep(10 * time.Second)
			}
			return nil
		}),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		// chromedp.Sleep(3142*time.Millisecond),
		chromedp.OuterHTML("html", &htmlContent),
	)
	if err != nil {
		return nil, nil
	}

	// fmt.Println(htmlContent)

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("error creating document\n%w", err)
	} else if doc == nil {
		return nil, fmt.Errorf("no document. check chrome browser exists\n%w", err)
	}

	return doc, nil
}

func selectAllMatched(doc *goquery.Document, cssPath string) []string {

	matched := make([]string, 0)
	// Find the elements by the CSS selector
	doc.Find(cssPath).Each(func(i int, s *goquery.Selection) {
		matched = append(matched, s.Text())
	})

	return matched
}

func selectAllMatchedWithChildPath(doc *goquery.Document, cssPath string, pathChild ...string) [][]string {

	matched := make([][]string, 0)
	for i := 0; i < len(pathChild); i++ {
		matched[i] = make([]string, 0)
	}
	// Find the elements by the CSS selector
	doc.Find(cssPath).Each(func(_ int, s *goquery.Selection) {
		for i, c := range pathChild {
			s.Find(c).Each(func(_ int, s2 *goquery.Selection) {
				matched[i] = append(matched[i], s2.Text())
			})
		}
	})

	return matched
}
