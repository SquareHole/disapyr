package internal

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"

	"github.com/mr-tron/base58"
)

// HideIdentifier encrypts the provided identifier using AES-GCM,
// prepends the nonce, and returns a Base58-encoded string.
func HideIdentifier(id string, key []byte) (string, error) {
	// AES-GCM requires a nonce. The recommended size is 12 bytes.
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Create a new AES cipher using the key.
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create the AES-GCM cipher mode instance.
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Encrypt the identifier.
	encrypted := aead.Seal(nil, nonce, []byte(id), nil)

	// Concatenate the nonce and the encrypted data.
	combined := append(nonce, encrypted...)

	// Encode the result using Base58 (Bitcoin alphabet).
	encoded := base58.Encode(combined)
	return encoded, nil
}
