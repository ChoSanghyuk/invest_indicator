package util

import (
	"encoding/base64"
	"fmt"
	"testing"
)

func TestBase64(t *testing.T) {

	e := base64.StdEncoding.EncodeToString([]byte(""))
	fmt.Println(e)
	d, err := base64.StdEncoding.DecodeString("")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(string(d))
}
