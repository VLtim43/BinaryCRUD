package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

const (
	keySize        = 2048
	privateKeyFile = "private.pem"
	publicKeyFile  = "public.pem"
)

// KeysDir is where RSA keys are stored
const KeysDir = "data/keys"

var (
	instance *RSACrypto
	once     sync.Once
	enabled  = true // Encryption enabled by default
	mu       sync.RWMutex
)

// RSACrypto handles RSA encryption/decryption
type RSACrypto struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

// GetInstance returns the singleton RSACrypto instance
func GetInstance() (*RSACrypto, error) {
	var initErr error
	once.Do(func() {
		instance = &RSACrypto{}
		initErr = instance.loadOrGenerateKeys()
	})
	if initErr != nil {
		return nil, initErr
	}
	return instance, nil
}

// loadOrGenerateKeys loads existing keys or generates new ones
func (r *RSACrypto) loadOrGenerateKeys() error {
	if err := os.MkdirAll(KeysDir, 0755); err != nil {
		return fmt.Errorf("failed to create keys directory: %w", err)
	}

	privatePath := filepath.Join(KeysDir, privateKeyFile)
	publicPath := filepath.Join(KeysDir, publicKeyFile)

	// Try to load existing keys
	if _, err := os.Stat(privatePath); err == nil {
		if err := r.loadKeys(privatePath, publicPath); err == nil {
			return nil
		}
	}

	// Generate new keys
	return r.generateAndSaveKeys(privatePath, publicPath)
}

// generateAndSaveKeys creates new RSA keys and saves them
func (r *RSACrypto) generateAndSaveKeys(privatePath, publicPath string) error {
	privateKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return fmt.Errorf("failed to generate RSA key: %w", err)
	}

	r.privateKey = privateKey
	r.publicKey = &privateKey.PublicKey

	// Save private key
	privateBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privatePEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateBytes,
	})
	if err := os.WriteFile(privatePath, privatePEM, 0600); err != nil {
		return fmt.Errorf("failed to save private key: %w", err)
	}

	// Save public key
	publicBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return fmt.Errorf("failed to marshal public key: %w", err)
	}
	publicPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicBytes,
	})
	if err := os.WriteFile(publicPath, publicPEM, 0644); err != nil {
		return fmt.Errorf("failed to save public key: %w", err)
	}

	return nil
}

// loadKeys loads existing RSA keys from files
func (r *RSACrypto) loadKeys(privatePath, publicPath string) error {
	// Load private key
	privateData, err := os.ReadFile(privatePath)
	if err != nil {
		return fmt.Errorf("failed to read private key: %w", err)
	}

	block, _ := pem.Decode(privateData)
	if block == nil {
		return fmt.Errorf("failed to decode private key PEM")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %w", err)
	}

	r.privateKey = privateKey
	r.publicKey = &privateKey.PublicKey

	return nil
}

// Encrypt encrypts plaintext using RSA-OAEP
func (r *RSACrypto) Encrypt(plaintext []byte) ([]byte, error) {
	if len(plaintext) == 0 {
		return []byte{}, nil
	}

	hash := sha256.New()
	ciphertext, err := rsa.EncryptOAEP(hash, rand.Reader, r.publicKey, plaintext, nil)
	if err != nil {
		return nil, fmt.Errorf("encryption failed: %w", err)
	}

	return ciphertext, nil
}

// Decrypt decrypts ciphertext using RSA-OAEP
func (r *RSACrypto) Decrypt(ciphertext []byte) ([]byte, error) {
	if len(ciphertext) == 0 {
		return []byte{}, nil
	}

	hash := sha256.New()
	plaintext, err := rsa.DecryptOAEP(hash, rand.Reader, r.privateKey, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return plaintext, nil
}

// EncryptString encrypts a string and returns encrypted bytes
// If encryption is disabled, returns the plaintext as bytes
func (r *RSACrypto) EncryptString(plaintext string) ([]byte, error) {
	if !IsEnabled() {
		return []byte(plaintext), nil
	}
	return r.Encrypt([]byte(plaintext))
}

// DecryptString decrypts bytes and returns the original string
// If encryption is disabled, returns the bytes as string directly
func (r *RSACrypto) DecryptString(ciphertext []byte) (string, error) {
	if !IsEnabled() {
		return string(ciphertext), nil
	}
	plaintext, err := r.Decrypt(ciphertext)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

// IsEnabled returns whether encryption is currently enabled
func IsEnabled() bool {
	mu.RLock()
	defer mu.RUnlock()
	return enabled
}

// SetEnabled enables or disables encryption
func SetEnabled(enable bool) {
	mu.Lock()
	defer mu.Unlock()
	enabled = enable
}

// Reset clears the singleton instance so new keys will be generated on next GetInstance call
func Reset() {
	mu.Lock()
	defer mu.Unlock()
	instance = nil
	once = sync.Once{}
}
