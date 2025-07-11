package utilsCrypto

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kms"
)

// DataKeyInfo represents information about a data key
type DataKeyInfo struct {
	OrganizationID    string
	EncryptedDataKey  string
	CreatedAt         time.Time
	LastRotatedAt     time.Time
	EncryptionContext map[string]string
}

// EnvelopeService manages data keys and provides encryption/decryption operations
type EnvelopeService struct {
	kmsClient    *kms.Client
	masterKeyID  string
	dataKeyStore map[string]*DataKeyInfo // In-memory store for data keys
	cache        map[string][]byte       // In-memory cache for plaintext data keys (use with caution)
	mu           sync.RWMutex            // Mutex for thread-safe access to dataKeyStore and cache
}

// NewEnvelopeService creates a new instance of EnvelopeService
func NewEnvelopeService(ctx context.Context, masterKeyID string) (*EnvelopeService, error) {
	// Create KMS client
	kmsClient, err := generateKmsClient()
	if err != nil {
		return nil, err
	}

	return &EnvelopeService{
		kmsClient:    kmsClient,
		masterKeyID:  masterKeyID,
		dataKeyStore: make(map[string]*DataKeyInfo),
		cache:        make(map[string][]byte),
	}, nil
}

// CreateDataKey generates a new data key for an organization
func (s *EnvelopeService) CreateDataKey(ctx context.Context, orgID string) (*DataKeyInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if a data key already exists for this organization
	if _, exists := s.dataKeyStore[orgID]; exists {
		return nil, errors.New("data key already exists for this organization")
	}

	// Create encryption context with organization ID
	encryptionContext := map[string]string{
		"organization_id": orgID,
		"created_at":      time.Now().UTC().Format(time.RFC3339),
	}

	// Generate a new data key
	dataKeyInput := &kms.GenerateDataKeyInput{
		KeyId:             aws.String(s.masterKeyID),
		EncryptionContext: encryptionContext,
		KeySpec:           "AES_256",
	}

	dataKeyOutput, err := s.kmsClient.GenerateDataKey(ctx, dataKeyInput)
	if err != nil {
		return nil, fmt.Errorf("failed to generate data key: %w", err)
	}

	// Encode the encrypted data key for storage
	encryptedDataKeyBase64 := base64.StdEncoding.EncodeToString(dataKeyOutput.CiphertextBlob)

	// Store the data key info in memory
	now := time.Now().UTC()
	dataKeyInfo := &DataKeyInfo{
		OrganizationID:    orgID,
		EncryptedDataKey:  encryptedDataKeyBase64,
		CreatedAt:         now,
		LastRotatedAt:     now,
		EncryptionContext: encryptionContext,
	}

	// Store in memory
	s.dataKeyStore[orgID] = dataKeyInfo

	// Optionally cache the plaintext key (with caution)
	// s.cache[orgID] = dataKeyOutput.Plaintext

	return dataKeyInfo, nil
}

// FetchDataKey retrieves and decrypts a data key for an organization
func (s *EnvelopeService) FetchDataKey(ctx context.Context, orgID string) ([]byte, error) {
	s.mu.RLock()
	// Check if we have the plaintext key in cache
	if plaintextKey, ok := s.cache[orgID]; ok {
		s.mu.RUnlock()
		return plaintextKey, nil
	}

	// Retrieve the data key info from memory
	dataKeyInfo, exists := s.dataKeyStore[orgID]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no data key found for organization: %s", orgID)
	}

	// Decode the base64-encoded encrypted data key
	encryptedDataKey, err := base64.StdEncoding.DecodeString(dataKeyInfo.EncryptedDataKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode encrypted data key: %w", err)
	}

	// Decrypt the data key using KMS
	decryptInput := &kms.DecryptInput{
		CiphertextBlob:    encryptedDataKey,
		EncryptionContext: dataKeyInfo.EncryptionContext,
		KeyId:             aws.String(s.masterKeyID),
	}

	decryptOutput, err := s.kmsClient.Decrypt(ctx, decryptInput)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data key: %w", err)
	}

	// Optionally cache the plaintext key (with caution)
	// Consider using a mutex when updating the cache
	s.mu.Lock()
	s.cache[orgID] = decryptOutput.Plaintext
	s.mu.Unlock()

	return decryptOutput.Plaintext, nil
}

// RotateDataKey generates a new data key for an organization
func (s *EnvelopeService) RotateDataKey(ctx context.Context, orgID string) (*DataKeyInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get the existing data key info
	dataKeyInfo, exists := s.dataKeyStore[orgID]
	if !exists {
		return nil, fmt.Errorf("no data key found for organization: %s", orgID)
	}

	// Update encryption context with rotation information
	encryptionContext := dataKeyInfo.EncryptionContext
	encryptionContext["rotated_at"] = time.Now().UTC().Format(time.RFC3339)

	// Generate a new data key
	dataKeyInput := &kms.GenerateDataKeyInput{
		KeyId:             aws.String(s.masterKeyID),
		EncryptionContext: encryptionContext,
		KeySpec:           "AES_256",
	}

	dataKeyOutput, err := s.kmsClient.GenerateDataKey(ctx, dataKeyInput)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new data key: %w", err)
	}

	// Encode the new encrypted data key
	encryptedDataKeyBase64 := base64.StdEncoding.EncodeToString(dataKeyOutput.CiphertextBlob)

	// Update the data key info
	dataKeyInfo.EncryptedDataKey = encryptedDataKeyBase64
	dataKeyInfo.LastRotatedAt = time.Now().UTC()
	dataKeyInfo.EncryptionContext = encryptionContext

	// Update in memory store
	s.dataKeyStore[orgID] = dataKeyInfo

	// Clear the cached plaintext key if it exists
	delete(s.cache, orgID)

	return dataKeyInfo, nil
}

// Encrypt encrypts data for a specific organization
func (s *EnvelopeService) Encrypt(ctx context.Context, orgID string, plaintext []byte) (string, error) {
	// Get the data key for the organization
	dataKey, err := s.FetchDataKey(ctx, orgID)
	if err != nil {
		return "", fmt.Errorf("failed to fetch data key: %w", err)
	}

	// Create a new AES cipher using the data key
	block, err := aes.NewCipher(dataKey)
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Create a GCM cipher mode
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Create a nonce (must be unique for each encryption)
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt the plaintext
	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)

	// Encode the ciphertext for storage or transmission
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts data for a specific organization
func (s *EnvelopeService) Decrypt(ctx context.Context, orgID string, ciphertextBase64 string) ([]byte, error) {
	// Get the data key for the organization
	dataKey, err := s.FetchDataKey(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data key: %w", err)
	}

	// Decode the base64-encoded ciphertext
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	// Create a new AES cipher using the data key
	block, err := aes.NewCipher(dataKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Create a GCM cipher mode
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Check if the ciphertext is valid
	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	// Extract the nonce and actual ciphertext
	nonce, encryptedData := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt the data
	plaintext, err := aesGCM.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt ciphertext: %w", err)
	}

	return plaintext, nil
}

// GetAllDataKeys returns all data keys stored in memory
func (s *EnvelopeService) GetAllDataKeys() map[string]*DataKeyInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Create a copy to avoid exposing internal data structure
	result := make(map[string]*DataKeyInfo, len(s.dataKeyStore))
	for orgID, keyInfo := range s.dataKeyStore {
		keyCopy := *keyInfo // Create a copy
		result[orgID] = &keyCopy
	}

	return result
}

// SerializeDataKeyStore serializes the data key store to JSON
func (s *EnvelopeService) SerializeDataKeyStore() ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return json.Marshal(s.dataKeyStore)
}

// DeserializeDataKeyStore loads the data key store from JSON
func (s *EnvelopeService) DeserializeDataKeyStore(data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return json.Unmarshal(data, &s.dataKeyStore)
}

// DeleteDataKey removes a data key from memory
func (s *EnvelopeService) DeleteDataKey(orgID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.dataKeyStore, orgID)
	delete(s.cache, orgID)
}
