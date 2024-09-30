package main

import (
	"crypto/aes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
)

// KeyStore represents the structure to store the encryption key and IV
type KeyStore struct {
	SecretKey string `json:"secret_key"`
	IV        string `json:"iv"`
}

// GenerateNewEncryptionKeys generates a new AES-256 key and IV
func GenerateNewEncryptionKeys() (string, string, error) {
	key := make([]byte, 32)           // AES-256 uses a 32-byte key
	iv := make([]byte, aes.BlockSize) // AES block size (16 bytes for AES)

	// Generate random bytes for key and IV
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return "", "", err
	}
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", "", err
	}

	// Base64 encode both key and IV for easy storage in a file
	encodedKey := base64.StdEncoding.EncodeToString(key)
	encodedIV := base64.StdEncoding.EncodeToString(iv)

	return encodedKey, encodedIV, nil
}

// SaveKeysToFile saves the secret key and IV to a file in JSON format
func SaveKeysToFile(filename string, keyStore KeyStore) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // Pretty-print JSON
	return encoder.Encode(keyStore)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: keygen <output_file>")
		os.Exit(1)
	}

	outputFile := os.Args[1]

	// Generate new encryption keys
	secretKey, iv, err := GenerateNewEncryptionKeys()
	if err != nil {
		log.Fatalf("Error generating encryption keys: %v", err)
	}

	// Prepare the KeyStore struct
	keyStore := KeyStore{
		SecretKey: secretKey,
		IV:        iv,
	}

	// Save the keys to the specified file
	err = SaveKeysToFile(outputFile, keyStore)
	if err != nil {
		log.Fatalf("Error saving keys to file: %v", err)
	}

	fmt.Printf("Encryption keys saved to %s\n", outputFile)
}

