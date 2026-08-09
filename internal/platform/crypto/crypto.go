// Package crypto реализует асимметричное шифрование тела запросов агент → сервер.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
)

// HeaderEncrypted — HTTP-заголовок, помечающий зашифрованное тело запроса.
const HeaderEncrypted = "X-Content-Encrypted"

// HeaderEncryptedValue — значение заголовка HeaderEncrypted: имя схемы шифрования.
const HeaderEncryptedValue = "rsa-aes256gcm"

const aesKeySize = 32

// LoadPublicKey читает RSA-публичный ключ из PEM-файла (PKIX или PKCS1).
func LoadPublicKey(path string) (*rsa.PublicKey, error) {
	block, err := readPEM(path)
	if err != nil {
		return nil, err
	}

	if key, err := x509.ParsePKIXPublicKey(block.Bytes); err == nil {
		rsaKey, ok := key.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("%s: ожидался RSA-ключ, получен %T", path, key)
		}

		return rsaKey, nil
	}

	key, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("%s: разбор публичного ключа: %w", path, err)
	}

	return key, nil
}

// LoadPrivateKey читает RSA-приватный ключ из PEM-файла (PKCS8 или PKCS1).
func LoadPrivateKey(path string) (*rsa.PrivateKey, error) {
	block, err := readPEM(path)
	if err != nil {
		return nil, err
	}

	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		rsaKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("%s: ожидался RSA-ключ, получен %T", path, key)
		}

		return rsaKey, nil
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("%s: разбор приватного ключа: %w", path, err)
	}

	return key, nil
}

// Encrypt шифрует data гибридной схемой: AES-256-GCM со случайным ключом,
// зашифрованным RSA-OAEP публичным ключом publicKey.
func Encrypt(publicKey *rsa.PublicKey, data []byte) ([]byte, error) {
	aesKey := make([]byte, aesKeySize)
	if _, err := rand.Read(aesKey); err != nil {
		return nil, fmt.Errorf("generate aes key: %w", err)
	}

	gcm, err := newGCM(aesKey)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("generate nonce: %w", err)
	}

	encryptedKey, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, aesKey, nil)
	if err != nil {
		return nil, fmt.Errorf("encrypt aes key: %w", err)
	}

	out := make([]byte, 0, len(encryptedKey)+len(nonce)+len(data)+gcm.Overhead())
	out = append(out, encryptedKey...)
	out = append(out, nonce...)
	out = gcm.Seal(out, nonce, data, nil)

	return out, nil
}

// Decrypt расшифровывает сообщение, созданное Encrypt, приватным ключом privateKey.
func Decrypt(privateKey *rsa.PrivateKey, data []byte) ([]byte, error) {
	keySize := privateKey.Size()

	if len(data) < keySize {
		return nil, errors.New("encrypted message too short")
	}

	aesKey, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, data[:keySize], nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt aes key: %w", err)
	}

	gcm, err := newGCM(aesKey)
	if err != nil {
		return nil, err
	}

	rest := data[keySize:]
	if len(rest) < gcm.NonceSize() {
		return nil, errors.New("encrypted message too short")
	}

	nonce, ciphertext := rest[:gcm.NonceSize()], rest[gcm.NonceSize():]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt payload: %w", err)
	}

	return plaintext, nil
}

func newGCM(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create aes cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create gcm: %w", err)
	}

	return gcm, nil
}

func readPEM(path string) (*pem.Block, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read key file: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("%s: не найден PEM-блок", path)
	}

	return block, nil
}
