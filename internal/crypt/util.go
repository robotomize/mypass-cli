package crypt

import (
	"crypto/sha512"
	"golang.org/x/crypto/pbkdf2"
)

func GeneratePrivateKeyFromPassword(keyLen int) func(password string) ([]byte, error) {
	return func(password string) ([]byte, error) {
		hash := sha512.New()
		hash.Write([]byte(password))

		key := pbkdf2.Key([]byte(password), hash.Sum(nil), 4096, keyLen, sha512.New)

		return key, nil
	}
}

func GeneratePrivateKeyAES() func(password string) ([]byte, error) {
	return GeneratePrivateKeyFromPassword(32)
}

func GeneratePrivateKeyDES() func(password string) ([]byte, error) {
	return GeneratePrivateKeyFromPassword(8)
}
