package util

import (
	"encoding/base64"
	"fmt"
)

func Decode(target *string) {

	decoded, err := base64.StdEncoding.DecodeString(*target)
	if err != nil {
		fmt.Println(err)
	}

	new := string(decoded)
	*target = new
}
