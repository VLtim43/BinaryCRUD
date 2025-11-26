package test

import (
	"BinaryCRUD/backend/crypto"
	"testing"
)

func TestRSAEncryptDecrypt(t *testing.T) {
	rsaCrypto, err := crypto.GetInstance()
	if err != nil {
		t.Fatalf("Failed to get RSA instance: %v", err)
	}

	testCases := []string{
		"Hello World",
		"John Doe",
		"Test Customer Name",
		"Ação com acentos",
		"",
	}

	for _, original := range testCases {
		encrypted, err := rsaCrypto.EncryptString(original)
		if err != nil {
			t.Errorf("Failed to encrypt '%s': %v", original, err)
			continue
		}

		decrypted, err := rsaCrypto.DecryptString(encrypted)
		if err != nil {
			t.Errorf("Failed to decrypt '%s': %v", original, err)
			continue
		}

		if decrypted != original {
			t.Errorf("Decrypted value mismatch: expected '%s', got '%s'", original, decrypted)
		}
	}
}

func TestRSAEncryptedDataIsDifferent(t *testing.T) {
	rsaCrypto, err := crypto.GetInstance()
	if err != nil {
		t.Fatalf("Failed to get RSA instance: %v", err)
	}

	original := "Secret Customer Name"
	encrypted, err := rsaCrypto.EncryptString(original)
	if err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	// Encrypted data should be different from original
	if string(encrypted) == original {
		t.Error("Encrypted data should be different from original")
	}

	// Encrypted data should be longer (RSA produces fixed-size output)
	if len(encrypted) <= len(original) {
		t.Error("Encrypted data should be larger than original for RSA")
	}
}
