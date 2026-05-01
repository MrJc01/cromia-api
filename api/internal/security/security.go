package security

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Configuração padrão do Argon2id
const (
	time    uint32 = 1
	memory  uint32 = 64 * 1024
	threads uint8  = 4
	keyLen  uint32 = 32
	saltLen uint32 = 16
)

// HashAPIKey gera o hash de uma API Key
func HashAPIKey(key string) (string, error) {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(key), salt, time, memory, uint8(threads), keyLen)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, memory, time, threads, b64Salt, b64Hash), nil
}

// CompareAPIKey verifica se a chave fornecida corresponde ao hash
func CompareAPIKey(key, hash string) (bool, error) {
	parts := strings.Split(hash, "$")
	if len(parts) != 6 {
		return false, fmt.Errorf("formato de hash inválido")
	}

	var version int
	var memory, time uint32
	var threads uint8
	var b64Salt, b64Hash string

	_, err := fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil {
		return false, err
	}
	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads)
	if err != nil {
		return false, err
	}
	b64Salt = parts[4]
	b64Hash = parts[5]

	salt, err := base64.RawStdEncoding.DecodeString(b64Salt)
	if err != nil {
		return false, err
	}

	hashToCompare := argon2.IDKey([]byte(key), salt, time, memory, uint8(threads), keyLen)
	computedHash := base64.RawStdEncoding.EncodeToString(hashToCompare)

	return computedHash == b64Hash, nil
}

func GenerateRandomString(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}
