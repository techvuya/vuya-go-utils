package utilsCrypto

import "testing"

func TestEncryptDecrypt(t *testing.T) {
	key := []byte("supersecretkey32byteslong1234567")
	plaintext := "Hello, World!"

	ciphertext, err := EncryptBytes(key, []byte(plaintext))
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}
	if string(ciphertext) == plaintext {
		t.Errorf("Encrypt: plaintext same as cipherText, got %s, want %s", ciphertext, plaintext)
	}

	decryptedPlaintext, err := DecryptBytes(key, ciphertext)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}
	if string(decryptedPlaintext) != plaintext {
		t.Errorf("Decrypt: plaintext mismatch, got %s, want %s", decryptedPlaintext, plaintext)
	}
}
