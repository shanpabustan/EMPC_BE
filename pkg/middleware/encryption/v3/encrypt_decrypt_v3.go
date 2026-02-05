package encrypDecryptV3

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
)

// createAESGCM initializes AES-GCM using the provided key.
func createAESGCM(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

// EncryptV2 encrypts the plaintext using AES-GCM.
func EncryptV3(plaintext, key []byte) ([]byte, []byte, error) {
	aesGCM, err := createAESGCM(key)
	if err != nil {
		return nil, nil, err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, err
	}

	return aesGCM.Seal(nil, nonce, plaintext, nil), nonce, nil
}

// DecryptV2 decrypts the ciphertext using AES-GCM.
func DecryptV3(ciphertext, key []byte) ([]byte, error) {
	aesGCM, err := createAESGCM(key)
	if err != nil {
		return nil, err
	}
	_, nonce, _ := EncryptV3(ciphertext, key)
	return aesGCM.Open(nil, nonce, ciphertext, nil)
}
