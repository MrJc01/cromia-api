package security

import (
	"crypto/rand"
	"encoding/base64"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"golang.org/x/crypto/bcrypt"
)

func HashAPIKey(key string) (string, error) {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:]), nil
}

func CompareAPIKey(key, hash string) (bool, error) {
	expectedHash := sha256.Sum256([]byte(key))
	expectedHashHex := hex.EncodeToString(expectedHash[:])
	return subtle.ConstantTimeCompare([]byte(hash), []byte(expectedHashHex)) == 1, nil
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
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func ComparePassword(password, hash string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil, err
}
