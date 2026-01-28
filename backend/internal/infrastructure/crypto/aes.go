// Package crypto provides encryption services for sensitive data.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

// Common errors for crypto operations.
var (
	ErrInvalidKeyLength = errors.New("encryption key must be exactly 32 bytes for AES-256")
	ErrInvalidKeyFormat = errors.New("encryption key must be valid base64")
	ErrDecryptionFailed = errors.New("decryption failed: invalid ciphertext or nonce")
	ErrInvalidNonce     = errors.New("nonce must be exactly 12 bytes for AES-256-GCM")
)

// CryptoService defines the interface for encryption and decryption operations.
type CryptoService interface {
	// Encrypt encrypts plaintext and returns ciphertext and nonce.
	// The nonce must be stored alongside the ciphertext for decryption.
	Encrypt(plaintext []byte) (ciphertext []byte, nonce []byte, err error)

	// Decrypt decrypts ciphertext using the provided nonce.
	Decrypt(ciphertext []byte, nonce []byte) (plaintext []byte, err error)
}

// AESCryptoService implements CryptoService using AES-256-GCM.
type AESCryptoService struct {
	gcm cipher.AEAD
}

// NewAESCryptoService creates a new AES-256-GCM crypto service.
// The key must be a base64-encoded 32-byte key.
func NewAESCryptoService(base64Key string) (*AESCryptoService, error) {
	// Decode the base64 key
	key, err := base64.StdEncoding.DecodeString(base64Key)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidKeyFormat, err)
	}

	// Validate key length (must be 32 bytes for AES-256)
	if len(key) != 32 {
		return nil, fmt.Errorf("%w: got %d bytes", ErrInvalidKeyLength, len(key))
	}

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM mode: %w", err)
	}

	return &AESCryptoService{gcm: gcm}, nil
}

// Encrypt encrypts plaintext using AES-256-GCM.
// Returns the ciphertext and a randomly generated 12-byte nonce.
// The nonce must be stored alongside the ciphertext for later decryption.
func (s *AESCryptoService) Encrypt(plaintext []byte) ([]byte, []byte, error) {
	// Generate a random nonce (12 bytes for GCM)
	nonce := make([]byte, s.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt the plaintext
	// Seal appends the encrypted data to the first argument (nil = new slice)
	ciphertext := s.gcm.Seal(nil, nonce, plaintext, nil)

	return ciphertext, nonce, nil
}

// Decrypt decrypts ciphertext using AES-256-GCM with the provided nonce.
func (s *AESCryptoService) Decrypt(ciphertext []byte, nonce []byte) ([]byte, error) {
	// Validate nonce length
	if len(nonce) != s.gcm.NonceSize() {
		return nil, fmt.Errorf("%w: got %d bytes, expected %d", ErrInvalidNonce, len(nonce), s.gcm.NonceSize())
	}

	// Decrypt the ciphertext
	plaintext, err := s.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
	}

	return plaintext, nil
}

// GenerateKey generates a new random 32-byte key and returns it as base64.
// This is a helper function for generating encryption keys.
func GenerateKey() (string, error) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return "", fmt.Errorf("failed to generate random key: %w", err)
	}
	return base64.StdEncoding.EncodeToString(key), nil
}
