package libgin

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
)

// Encrypt an array of bytes with AES.
func Encrypt(key, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)
	return ciphertext, nil
}

// EncryptURLString encrypts a string using AES and returns it in URL encoded base64.
func EncryptURLString(key []byte, plaintext string) (string, error) {
	cipherbytes, err := Encrypt(key, []byte(plaintext))
	if err != nil {
		return "", err
	}
	// convert to URL encoded base64
	return base64.URLEncoding.EncodeToString(cipherbytes), nil
}

// EncryptString encrypts a string using AES and returns it in base64.
func EncryptString(key []byte, plaintext string) (string, error) {
	cipherbytes, err := Encrypt(key, []byte(plaintext))
	if err != nil {
		return "", err
	}
	// convert to (standard) encoded base64
	return base64.StdEncoding.EncodeToString(cipherbytes), nil
}

// Decrypt an AES encrypted array of bytes.
func Decrypt(key, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	if len(ciphertext) < aes.BlockSize {
		return nil, err
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)

	// XORKeyStream can work in-place if the two arguments are the same.
	stream.XORKeyStream(ciphertext, ciphertext)

	return ciphertext, nil
}

// DecryptURLString decrypts an AES encrypted URL encoded base64 string.
func DecryptURLString(key []byte, encstring string) (string, error) {
	ciphertext, err := base64.URLEncoding.DecodeString(encstring)
	if err != nil {
		return "", err
	}
	plainbytes, err := Decrypt(key, ciphertext)
	if err != nil {
		return "", err
	}
	return string(plainbytes), nil
}

// DecryptString decrypts an AES encrypted base64 string.
func DecryptString(key []byte, encstring string) (string, error) {
	ciphertext, _ := base64.StdEncoding.DecodeString(encstring)
	plainbytes, err := Decrypt(key, ciphertext)
	if err != nil {
		return "", err
	}
	return string(plainbytes), nil
}
