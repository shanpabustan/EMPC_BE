package hashingV1

import (
	"crypto/sha512"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"
)

// Hash string ...
func GenerateHash(text string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(text), 14)
	return string(bytes), err
}

// Validate hash string...
func ValidateHash(text, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(text))
	return err == nil
}

func HashDataSHA512(data string) string {
	hash := sha512.Sum512([]byte(data))
	return hex.EncodeToString(hash[:])
}

func ValidateHashSHA512(input, storedHash string) bool {
	computedHash := HashDataSHA512(input)
	return computedHash == storedHash
}
