package util

import (
	"fmt"
	"testing"
	"time"

	"github.com/robfig/cron"
)

func TestCron(t *testing.T) {
	c := cron.New()
	c.AddFunc("* * * * * *", func() {
		fmt.Println("Hello")
	})

	c.Start()
	time.Sleep(time.Minute * 3)
}
