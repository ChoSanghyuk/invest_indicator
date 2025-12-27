package util

import (
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

func Hex2Bytes(str string) []byte {
	if strings.HasPrefix(str, "0x") {
		str, _ = strings.CutPrefix(str, "0x")
	}

	return common.Hex2Bytes(str)
}
