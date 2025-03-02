package util

import (
	"os"
	"testing"
)

/*
export key=""
export plain=""
export cipher=""
지정 필요

이후 같은 터미널에서 go test ./config -run {함수명} -v 수행

*/

func TestEncrypt(t *testing.T) {

	key := os.Getenv("key")
	plain := os.Getenv("plain")

	encrypted, err := Encrypt([]byte(key), plain)
	if err != nil {
		t.Error(err)
	}
	t.Logf("encrypted: %s\n", encrypted)
}

func TestDecrypt(t *testing.T) {

	key := os.Getenv("key")
	cipher := os.Getenv("cipher")

	decrypted, err := Decrypt([]byte(key), cipher)
	if err != nil {
		t.Error(err)
	}
	t.Logf("decrypted: %s\n", decrypted)
}
