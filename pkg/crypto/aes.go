package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
)

const DataSourceKey = "c6659d5707bb63ee51d4b45c8c90bc73"

// pkcs7Padding pads a slice of bytes to a specified block size, using PKCS#7 padding.
func pkcs7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

// pkcs7Unpadding removes the PKCS#7 padding from a slice of bytes.
func pkcs7Unpadding(data []byte) ([]byte, error) {
	length := len(data)
	unpadding := int(data[length-1])
	if unpadding > length {
		return nil, errors.New("crypto/cipher: invalid padding")
	}
	return data[:(length - unpadding)], nil
}

func EncryptAES(key string, plaintext string) (string, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	ciphertext := pkcs7Padding([]byte(plaintext), block.BlockSize())

	tv, err := randomIV(block.BlockSize())
	if err != nil {
		return "", err
	}
	ciphertext = append(tv, ciphertext...)

	// CBC mode always works in whole blocks.
	if len(ciphertext)%aes.BlockSize != 0 {
		panic("ciphertext is not a multiple of the block size")
	}

	mode := cipher.NewCBCEncrypter(block, ciphertext[:aes.BlockSize])
	mode.CryptBlocks(ciphertext[aes.BlockSize:], ciphertext[aes.BlockSize:])

	return hex.EncodeToString(ciphertext), nil
}

func DecryptAES(key string, ciphertext string) (string, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}
	encodedCiphertext, err := hex.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}
	iv := encodedCiphertext[:aes.BlockSize]
	encodedCiphertext = encodedCiphertext[aes.BlockSize:]

	if len(encodedCiphertext)%aes.BlockSize != 0 {
		return "", errors.New("ciphertext is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(encodedCiphertext, encodedCiphertext)

	plaintext, err := pkcs7Unpadding(encodedCiphertext)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// randomIV generates a random AES initialization vector.
func randomIV(blockSize int) ([]byte, error) {
	iv := make([]byte, blockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	return iv, nil
}
