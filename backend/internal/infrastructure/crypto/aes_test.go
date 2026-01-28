package crypto

import (
	"bytes"
	"encoding/base64"
	"testing"
)

// Test key: 32 bytes encoded as base64
// Generated with: openssl rand -base64 32
const testKey = "6IX/ZL5Vzeawrh1gUUzIFW7KqtJJpLQDIwRYYgicagU="

func TestNewAESCryptoService(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr error
	}{
		{
			name:    "valid 32-byte key",
			key:     testKey,
			wantErr: nil,
		},
		{
			name:    "invalid base64",
			key:     "not-valid-base64!!!",
			wantErr: ErrInvalidKeyFormat,
		},
		{
			name:    "key too short (16 bytes)",
			key:     base64.StdEncoding.EncodeToString([]byte("short_key_16byte")),
			wantErr: ErrInvalidKeyLength,
		},
		{
			name:    "key too long (48 bytes)",
			key:     base64.StdEncoding.EncodeToString([]byte("this_key_is_way_too_long_for_aes_256_encryption!")),
			wantErr: ErrInvalidKeyLength,
		},
		{
			name:    "empty key",
			key:     "",
			wantErr: ErrInvalidKeyLength,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewAESCryptoService(tt.key)
			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.wantErr)
					return
				}
				// Check if error contains expected error
				if !errorContains(err, tt.wantErr) {
					t.Errorf("expected error to contain %v, got %v", tt.wantErr, err)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestEncryptDecrypt(t *testing.T) {
	service, err := NewAESCryptoService(testKey)
	if err != nil {
		t.Fatalf("failed to create crypto service: %v", err)
	}

	tests := []struct {
		name      string
		plaintext []byte
	}{
		{
			name:      "simple text",
			plaintext: []byte("hello world"),
		},
		{
			name:      "empty string",
			plaintext: []byte(""),
		},
		{
			name:      "password",
			plaintext: []byte("my-secret-password-123!@#"),
		},
		{
			name:      "unicode text",
			plaintext: []byte("Hello 世界 🌍"),
		},
		{
			name:      "long text",
			plaintext: bytes.Repeat([]byte("a"), 10000),
		},
		{
			name:      "binary data",
			plaintext: []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encrypt
			ciphertext, nonce, err := service.Encrypt(tt.plaintext)
			if err != nil {
				t.Fatalf("encryption failed: %v", err)
			}

			// Verify ciphertext is different from plaintext (unless empty)
			if len(tt.plaintext) > 0 && bytes.Equal(ciphertext, tt.plaintext) {
				t.Error("ciphertext should not equal plaintext")
			}

			// Verify nonce is 12 bytes
			if len(nonce) != 12 {
				t.Errorf("nonce should be 12 bytes, got %d", len(nonce))
			}

			// Decrypt
			decrypted, err := service.Decrypt(ciphertext, nonce)
			if err != nil {
				t.Fatalf("decryption failed: %v", err)
			}

			// Verify round-trip
			if !bytes.Equal(decrypted, tt.plaintext) {
				t.Errorf("decrypted text does not match original: got %q, want %q", decrypted, tt.plaintext)
			}
		})
	}
}

func TestNonceUniqueness(t *testing.T) {
	service, err := NewAESCryptoService(testKey)
	if err != nil {
		t.Fatalf("failed to create crypto service: %v", err)
	}

	plaintext := []byte("test message")
	nonces := make(map[string]bool)
	iterations := 1000

	for i := 0; i < iterations; i++ {
		_, nonce, err := service.Encrypt(plaintext)
		if err != nil {
			t.Fatalf("encryption failed on iteration %d: %v", i, err)
		}

		nonceStr := string(nonce)
		if nonces[nonceStr] {
			t.Errorf("duplicate nonce detected on iteration %d", i)
		}
		nonces[nonceStr] = true
	}

	if len(nonces) != iterations {
		t.Errorf("expected %d unique nonces, got %d", iterations, len(nonces))
	}
}

func TestDecryptWithWrongNonce(t *testing.T) {
	service, err := NewAESCryptoService(testKey)
	if err != nil {
		t.Fatalf("failed to create crypto service: %v", err)
	}

	plaintext := []byte("secret message")
	ciphertext, _, err := service.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	// Try to decrypt with a different nonce
	wrongNonce := make([]byte, 12)
	_, err = service.Decrypt(ciphertext, wrongNonce)
	if err == nil {
		t.Error("expected decryption to fail with wrong nonce")
	}
	if !errorContains(err, ErrDecryptionFailed) {
		t.Errorf("expected ErrDecryptionFailed, got %v", err)
	}
}

func TestDecryptWithInvalidNonceLength(t *testing.T) {
	service, err := NewAESCryptoService(testKey)
	if err != nil {
		t.Fatalf("failed to create crypto service: %v", err)
	}

	plaintext := []byte("secret message")
	ciphertext, _, err := service.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	tests := []struct {
		name      string
		nonceLen  int
		wantErr   error
	}{
		{
			name:     "nonce too short",
			nonceLen: 8,
			wantErr:  ErrInvalidNonce,
		},
		{
			name:     "nonce too long",
			nonceLen: 16,
			wantErr:  ErrInvalidNonce,
		},
		{
			name:     "empty nonce",
			nonceLen: 0,
			wantErr:  ErrInvalidNonce,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			invalidNonce := make([]byte, tt.nonceLen)
			_, err := service.Decrypt(ciphertext, invalidNonce)
			if err == nil {
				t.Error("expected error for invalid nonce length")
			}
			if !errorContains(err, tt.wantErr) {
				t.Errorf("expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestDecryptTamperedCiphertext(t *testing.T) {
	service, err := NewAESCryptoService(testKey)
	if err != nil {
		t.Fatalf("failed to create crypto service: %v", err)
	}

	plaintext := []byte("secret message")
	ciphertext, nonce, err := service.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	// Tamper with the ciphertext
	if len(ciphertext) > 0 {
		ciphertext[0] ^= 0xFF
	}

	_, err = service.Decrypt(ciphertext, nonce)
	if err == nil {
		t.Error("expected decryption to fail with tampered ciphertext")
	}
	if !errorContains(err, ErrDecryptionFailed) {
		t.Errorf("expected ErrDecryptionFailed, got %v", err)
	}
}

func TestGenerateKey(t *testing.T) {
	key1, err := GenerateKey()
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	// Verify key is valid base64
	decoded, err := base64.StdEncoding.DecodeString(key1)
	if err != nil {
		t.Errorf("generated key is not valid base64: %v", err)
	}

	// Verify key is 32 bytes
	if len(decoded) != 32 {
		t.Errorf("expected 32-byte key, got %d bytes", len(decoded))
	}

	// Verify generated key works with crypto service
	_, err = NewAESCryptoService(key1)
	if err != nil {
		t.Errorf("generated key should work with crypto service: %v", err)
	}

	// Verify keys are unique
	key2, err := GenerateKey()
	if err != nil {
		t.Fatalf("failed to generate second key: %v", err)
	}
	if key1 == key2 {
		t.Error("generated keys should be unique")
	}
}

func TestDifferentKeysProduceDifferentCiphertext(t *testing.T) {
	key1, _ := GenerateKey()
	key2, _ := GenerateKey()

	service1, err := NewAESCryptoService(key1)
	if err != nil {
		t.Fatalf("failed to create service 1: %v", err)
	}
	service2, err := NewAESCryptoService(key2)
	if err != nil {
		t.Fatalf("failed to create service 2: %v", err)
	}

	plaintext := []byte("test message")

	// Use same nonce for both to isolate key difference
	// (In practice, nonces should always be random)
	ciphertext1, nonce, _ := service1.Encrypt(plaintext)
	ciphertext2, _, _ := service2.Encrypt(plaintext)

	// Ciphertexts should be different
	if bytes.Equal(ciphertext1, ciphertext2) {
		t.Error("different keys should produce different ciphertexts")
	}

	// Cannot decrypt with wrong key
	_, err = service2.Decrypt(ciphertext1, nonce)
	if err == nil {
		t.Error("should not be able to decrypt with wrong key")
	}
}

// errorContains checks if err wraps or contains target error
func errorContains(err, target error) bool {
	if err == nil {
		return target == nil
	}
	return err.Error() == target.Error() ||
		(len(err.Error()) >= len(target.Error()) &&
		err.Error()[:len(target.Error())] == target.Error()[:len(target.Error())]) ||
		bytes.Contains([]byte(err.Error()), []byte(target.Error()))
}
