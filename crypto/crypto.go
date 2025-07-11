// Package crypto provides utility functions for encrypting and decrypting byte slices using XChaCha20-Poly1305.
package utilsCrypto

import (
	"crypto/rand"
	"fmt"

	"golang.org/x/crypto/chacha20poly1305"
)

// EncryptBytes encrypts the given plaintext with a 32-byte key using the XChaCha20-Poly1305 AEAD cipher.
// The returned byte slice contains the randomly generated nonce followed by the ciphertext and authentication tag.
//
// Parameters:
//   - key: A 32-byte secret key for encryption. It must be of length chacha20poly1305.KeySizeX.
//   - plainData: The plaintext data to encrypt.
//
// Returns:
//   - []byte: The nonce-prepended ciphertext.
//   - error: An error if key initialization or encryption fails.
func EncryptBytes(key []byte, plainData []byte) ([]byte, error) {
	// Create an AEAD instance for XChaCha20-Poly1305 using the provided key.
	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create encryption AEAD: %w", err)
	}

	// Allocate a nonce of the required size and fill it with cryptographically secure random bytes.
	nonce := make([]byte, aead.NonceSize(), aead.NonceSize()+len(plainData)+aead.Overhead())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Seal encrypts and authenticates the plaintext, appending the result to the nonce slice.
	encryptedMsg := aead.Seal(nonce, nonce, plainData, nil)

	return encryptedMsg, nil
}

// DecryptBytes decrypts data produced by EncryptBytes using the XChaCha20-Poly1305 AEAD cipher and the same 32-byte key.
// The input must contain the nonce as the first bytes, followed by the ciphertext and authentication tag.
//
// Parameters:
//   - key: A 32-byte secret key for decryption. Must match the key used for encryption.
//   - encryptedData: The nonce-prepended ciphertext returned by EncryptBytes.
//
// Returns:
//   - []byte: The decrypted plaintext.
//   - error: An error if key initialization, data length, or authentication fails.
func DecryptBytes(key []byte, encryptedData []byte) ([]byte, error) {
	// Initialize the AEAD instance with the provided key.
	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create decryption AEAD: %w", err)
	}

	// Ensure the input is at least as long as the nonce size.
	if len(encryptedData) < aead.NonceSize() {
		return nil, fmt.Errorf("ciphertext is too short: %d < %d", len(encryptedData), aead.NonceSize())
	}

	// Extract the nonce and the actual ciphertext.
	nonce := encryptedData[:aead.NonceSize()]
	ciphertext := encryptedData[aead.NonceSize():]

	// Open decrypts and authenticates the ciphertext, returning the plaintext or an error.
	plainData, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt or authenticate data: %w", err)
	}

	return plainData, nil
}
