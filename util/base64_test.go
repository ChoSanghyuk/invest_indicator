package util

import (
	"encoding/base64"
	"fmt"
	"os"
	"testing"
)

func TestBase64(t *testing.T) {

	input := os.Getenv("base64in")
	e := base64.StdEncoding.EncodeToString([]byte(input))
	fmt.Println(e)

	output := os.Getenv("base64out")
	d, err := base64.StdEncoding.DecodeString(output)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(string(d))
}
