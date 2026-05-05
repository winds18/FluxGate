package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
)

const tokenVaultVersion = "v1"

func EncryptTokenSecret(secret, token string) (string, error) {
	secret = strings.TrimSpace(secret)
	token = strings.TrimSpace(token)
	if secret == "" {
		return "", errors.New("token secret is required")
	}
	if token == "" {
		return "", errors.New("token is required")
	}
	gcm, err := tokenVaultGCM(secret)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nil, nonce, []byte(token), nil)
	return tokenVaultVersion + ":" +
		base64.RawURLEncoding.EncodeToString(nonce) + ":" +
		base64.RawURLEncoding.EncodeToString(ciphertext), nil
}

func DecryptTokenSecret(secret, encrypted string) (string, error) {
	secret = strings.TrimSpace(secret)
	encrypted = strings.TrimSpace(encrypted)
	if secret == "" {
		return "", errors.New("token secret is required")
	}
	parts := strings.Split(encrypted, ":")
	if len(parts) != 3 || parts[0] != tokenVaultVersion {
		return "", fmt.Errorf("unsupported encrypted token format")
	}
	nonce, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", err
	}
	ciphertext, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return "", err
	}
	gcm, err := tokenVaultGCM(secret)
	if err != nil {
		return "", err
	}
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func tokenVaultGCM(secret string) (cipher.AEAD, error) {
	key := sha256.Sum256([]byte("fluxgate-token-vault:" + secret))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}
