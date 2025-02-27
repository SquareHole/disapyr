package internal

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/mr-tron/base58"
	"github.com/stretchr/testify/assert"
)

func TestHideIdentifier(t *testing.T) {
	key := []byte("example key 1234") // 16 bytes key for AES-128

	t.Run("successful encryption", func(t *testing.T) {
		id := "test-identifier"
		encoded, err := HideIdentifier(id, key)
		assert.NoError(t, err)
		assert.NotEmpty(t, encoded)

		// Decode the Base58 encoded string
		decoded, err := base58.Decode(encoded)
		assert.NoError(t, err)

		// Extract the nonce and the encrypted data
		nonce := decoded[:12]
		encrypted := decoded[12:]

		// Create a new AES cipher using the key
		block, err := aes.NewCipher(key)
		assert.NoError(t, err)

		// Create the AES-GCM cipher mode instance
		aead, err := cipher.NewGCM(block)
		assert.NoError(t, err)

		// Decrypt the data
		decrypted, err := aead.Open(nil, nonce, encrypted, nil)
		assert.NoError(t, err)
		assert.Equal(t, id, string(decrypted))
	})

	t.Run("error generating nonce", func(t *testing.T) {
		// Override rand.Reader to return an error
		oldRandReader := rand.Reader
		defer func() { rand.Reader = oldRandReader }()
		rand.Reader = &errorReader{}

		id := "test-identifier"
		_, err := HideIdentifier(id, key)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to generate nonce")
	})

	t.Run("error creating AES cipher", func(t *testing.T) {
		invalidKey := []byte("short key")
		id := "test-identifier"
		_, err := HideIdentifier(id, invalidKey)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create cipher")
	})

	t.Run("error creating GCM", func(t *testing.T) {
		// Use a wrapper function to mock aes.NewCipher
		oldNewCipher := newCipher
		defer func() { newCipher = oldNewCipher }()
		newCipher = func(key []byte) (cipher.Block, error) {
			return nil, fmt.Errorf("failed to create cipher")
		}

		id := "test-identifier"
		_, err := newCipher(key)
		if err != nil {
			assert.Error(t, err, "expected error")
			assert.Contains(t, err.Error(), "failed to create cipher")
			return
		}
		encoded, err := HideIdentifier(id, key)
		assert.NoError(t, err)
		assert.NotEmpty(t, encoded)
	})
}

// Wrapper function for aes.NewCipher to allow mocking
var newCipher = aes.NewCipher

type errorReader struct{}

func (r *errorReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("random reader error")
}
