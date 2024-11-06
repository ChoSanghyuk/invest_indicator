package scrape

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/chromedp"
)

func crwalByChromedp() {
	ctx, cancel := chromedp.NewContext(
		context.Background(),
	)
	// to release the browser resources when
	// it is no longer needed
	defer cancel()

	var html string
	err := chromedp.Run(ctx,
		// visit the target page
		chromedp.Navigate("https://www.oecd.org/en/data/indicators/composite-leading-indicator-cli.html?oecdcontrol-b2a0dbca4d-var3=2008-01&oecdcontrol-b2a0dbca4d-var4=2024-12"),
		// wait for the page to load
		chromedp.Sleep(20000*time.Millisecond),
		// extract the raw HTML from the page
		chromedp.ActionFunc(func(ctx context.Context) error {
			// select the root node on the page
			rootNode, err := dom.GetDocument().Do(ctx)
			if err != nil {
				return err
			}
			html, err = dom.GetOuterHTML().WithNodeID(rootNode.NodeID).Do(ctx)
			return err
		}),
	)
	if err != nil {
		log.Fatal("Error while performing the automation logic:", err)
	}

	fmt.Println(html)
}
