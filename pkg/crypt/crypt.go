package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"errors"
	"fmt"
)

// EncryptWithPublicKey encrypts data with public RSA and AES.
func EncryptWithPublicKey(pubKey *rsa.PublicKey, data []byte) ([]byte, error) {
	aesKey := make([]byte, 32)
	if _, err := rand.Read(aesKey); err != nil {
		return nil, fmt.Errorf("failed to generate AES key: %w", err)
	}
	encryptedKey, err := rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		pubKey,
		aesKey,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt AES key: %w", err)
	}

	encryptedData, err := encryptWithAES(aesKey, data)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt data: %w", err)
	}

	res := make([]byte, len(encryptedKey)+len(encryptedData))
	copy(res[:len(encryptedKey)], encryptedKey)
	copy(res[len(encryptedKey):], encryptedData)

	return res, nil
}

// DecryptWithPrivateKey decrypts data with RSA and AES.
func DecryptWithPrivateKey(privKey *rsa.PrivateKey, data []byte) ([]byte, error) {
	keySize := privKey.Size()
	if len(data) < keySize {
		return nil, errors.New("ciphertext too short")
	}
	encryptedKey := data[:keySize]
	data = data[keySize:]

	aesKey, err := rsa.DecryptOAEP(
		sha256.New(),
		rand.Reader,
		privKey,
		encryptedKey,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt AES key: %w", err)
	}

	return decryptWithAES(aesKey, data)
}

// encryptWithAES encrypts data with AES-GCM.
func encryptWithAES(key, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	return gcm.Seal(nonce, nonce, data, nil), nil
}

// decryptWithAES decrypts data with AES-GCM.
func decryptWithAES(key, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}
