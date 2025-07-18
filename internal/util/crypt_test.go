package util

import (
	"fmt"
	"os"
	"testing"

	"golang.org/x/crypto/bcrypt"
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

func TestEncryptPassword(t *testing.T) {

	password := os.Getenv("password")
	fmt.Println(password)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Error(err)
	}
	t.Logf("encrypted password: %s\n", hashedPassword)

}
