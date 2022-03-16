package crypt

import (
	"crypto/sha512"
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/pbkdf2"
)

func GeneratePrivateKeyFromPassword(password string) ([]byte, error) {
	salt, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return nil, fmt.Errorf("bcrypt.GenerateFromPassword")
	}

	key := pbkdf2.Key([]byte(password), salt, 4096, 32, sha512.New)

	return key, nil
}
