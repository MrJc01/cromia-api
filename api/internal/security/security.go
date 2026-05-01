package security

import (
	"crypto/rand"
	"encoding/base64"
	"golang.org/x/crypto/bcrypt"
)

func HashAPIKey(key string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(key), bcrypt.DefaultCost)
	return string(bytes), err
}

func CompareAPIKey(key, hash string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(key))
	return err == nil, err
}

func GenerateRandomString(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func GenerateAPIKey() (string, string, error) {
	key := "crom_sk_" + GenerateRandomString(32)
	hash, err := HashAPIKey(key)
	return key, hash, err
}

func HashPassword(password string) (string, error) {
	return HashAPIKey(password)
}
