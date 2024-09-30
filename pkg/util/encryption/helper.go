package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
)

// Encrypt with the provided AES key and IV
func Encrypt(plainText, key, iv []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	cipherText := make([]byte, len(plainText))
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText, plainText)

	return base64.StdEncoding.EncodeToString(cipherText), nil
}

// Decrypt decrypts the base64-encoded cipher text using the provided key and IV
func Decrypt(encryptedData string, key, iv []byte) (string, error) {
	// Decode the Base64-encoded encrypted data
	cipherText, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return "", err
	}

	// Create AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// Check that the ciphertext length is a multiple of the block size
	if len(cipherText)%aes.BlockSize != 0 {
		return "", errors.New("ciphertext is not a multiple of the block size")
	}

	// Create CBC mode decrypter
	mode := cipher.NewCBCDecrypter(block, iv)

	// Decrypt the data
	decrypted := make([]byte, len(cipherText))
	mode.CryptBlocks(decrypted, cipherText)

	// Remove PKCS7 padding
	decrypted, err = pkcs7Unpad(decrypted, aes.BlockSize)
	if err != nil {
		return "", err
	}

	return string(decrypted), nil
}

// PKCS7 Unpadding
func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, errors.New("data is empty")
	}
	if length%blockSize != 0 {
		return nil, errors.New("data is not a multiple of the block size")
	}

	paddingLen := int(data[length-1])
	if paddingLen == 0 || paddingLen > blockSize {
		return nil, errors.New("invalid padding length")
	}

	// Verify padding bytes
	for i := length - paddingLen; i < length; i++ {
		if data[i] != byte(paddingLen) {
			return nil, errors.New("invalid padding")
		}
	}

	return data[:length-paddingLen], nil
}
