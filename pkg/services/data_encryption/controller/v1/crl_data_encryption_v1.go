package crlDataEncryptionV1

import (
	"crypto/rand"
	helper "golang-template-v3.1/pkg/global/json_response"
	encrypDecryptV1 "golang-template-v3.1/pkg/middleware/encryption/v1"
	mdlDataEncryptionV1 "golang-template-v3.1/pkg/services/data_encryption/model/v1"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"github.com/gofiber/fiber/v3"
)

func EncrypDecryptV1(c fiber.Ctx) error {
	dbData := &mdlDataEncryptionV1.DatabaseData{}
	if parseErr := c.Bind().Body(dbData); parseErr != nil {
		return helper.JSONResponseWithDataV1(c, "400", "failed to parse request body", dbData, fiber.StatusBadRequest)
	}

	// Start encrypting the data
	// check if secret key is provided, if not generate a new one
	if dbData.SecretKey == "" {
		secretKey, genErr := generateKey256Hex()
		if genErr != nil {
			return helper.JSONResponseWithErrorV1(c, "400", "failed to generate secret key", genErr, fiber.StatusBadRequest)
		}
		dbData.SecretKey = secretKey
		fmt.Println("secretKey:", dbData.SecretKey)
	}

	// encrypt the data
	encryptedDBHost, encErr := encrypDecryptV1.EncryptV1(dbData.DBHost, dbData.SecretKey)
	encryptedDBName, encErr := encrypDecryptV1.EncryptV1(dbData.DBName, dbData.SecretKey)
	encryptedDBUser, encErr := encrypDecryptV1.EncryptV1(dbData.DBUser, dbData.SecretKey)
	encryptedDBPass, encErr := encrypDecryptV1.EncryptV1(dbData.DBPass, dbData.SecretKey)

	if encErr != nil {
		return helper.JSONResponseWithErrorV1(c, "400", "failed to encrypt data", encErr, fiber.StatusBadRequest)
	}

	// pass the data to encryption function
	dbData.DBHost = encryptedDBHost
	dbData.DBName = encryptedDBName
	dbData.DBUser = encryptedDBUser
	dbData.DBPass = encryptedDBPass

	return helper.JSONResponseWithDataV1(c, "200", "success", dbData, fiber.StatusOK)
}

func DecryptDataV1(c fiber.Ctx) error {
	dbData := &mdlDataEncryptionV1.DatabaseData{}
	parseErr := c.Bind().Body(dbData)
	if parseErr != nil {
		return helper.JSONResponseWithErrorV1(c, "400", "failed to parse request body", parseErr, fiber.StatusBadRequest)
	}

	// null checking
	if dbData.SecretKey == "" {
		return helper.JSONResponseV1(c, "400", "secret key is required", fiber.StatusBadRequest)
	}

	// decrypt data
	decryptedHost, dErr := encrypDecryptV1.DecryptV1(dbData.DBHost, dbData.SecretKey)
	decryptedName, dErr := encrypDecryptV1.DecryptV1(dbData.DBName, dbData.SecretKey)
	decryptedUser, dErr := encrypDecryptV1.DecryptV1(dbData.DBUser, dbData.SecretKey)
	decryptedPass, dErr := encrypDecryptV1.DecryptV1(dbData.DBPass, dbData.SecretKey)
	if dErr != nil {
		return helper.JSONResponseWithErrorV1(c, "400", "failed to decrypt data", dErr, fiber.StatusBadRequest)
	}

	// pass the data to decryption function
	dbData.DBHost = decryptedHost
	dbData.DBName = decryptedName
	dbData.DBUser = decryptedUser
	dbData.DBPass = decryptedPass

	return helper.JSONResponseWithDataV1(c, "200", "success", dbData, fiber.StatusOK)
}

// GenerateKey256Bytes returns a securely-generated random 32-byte slice (256 bits).
// The caller is responsible for zeroing the slice if needed after use.
func generateKey256Bytes() ([]byte, error) {
	key := make([]byte, 16) // 32 bytes * 8 = 256 bits
	_, err := rand.Read(key)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random key: %w", err)
	}
	return key, nil
}

// GenerateKey256Hex returns the key encoded as a lowercase hexadecimal string (64 chars).
func generateKey256Hex() (string, error) {
	b, err := generateKey256Bytes()
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// GenerateKey256Base64 returns the key encoded in standard base64 (padding included).
func generateKey256Base64() (string, error) {
	b, err := generateKey256Bytes()
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}
