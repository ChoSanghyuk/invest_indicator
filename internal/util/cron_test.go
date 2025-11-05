package util

import (
	"fmt"
	"testing"
	"time"

	"github.com/robfig/cron"
	"github.com/rs/zerolog"
)

func TestCron(t *testing.T) {
	c := cron.New()
	c.AddFunc("* * * * * *", func() {
		fmt.Println("Hello")
	})

	c.Start()
	time.Sleep(time.Minute * 3)
}

func TestCron2(t *testing.T) {
	level, err := zerolog.ParseLevel("")
	fmt.Printf("%v\n", level)
	fmt.Printf("%v\n", err)
}
