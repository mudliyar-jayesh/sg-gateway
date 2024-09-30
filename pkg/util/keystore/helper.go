package utils

import (
	"encoding/json"
	"os"
)

// KeyStore represents the structure of the encryption key and IV stored in the JSON file
type KeyStore struct {
	SecretKey string `json:"secret_key"`
	IV        string `json:"iv"`
}

// LoadKeyStore loads the secret key and IV from the JSON key file
func LoadKeyStore(filePath string) (*KeyStore, error) {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var keyStore KeyStore
	if err := json.Unmarshal(file, &keyStore); err != nil {
		return nil, err
	}

	return &keyStore, nil
}
