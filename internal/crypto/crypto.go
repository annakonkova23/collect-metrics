package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"os"
)

func ReadPublicKey(filename string) (*rsa.PublicKey, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, fmt.Errorf("не удалось декодировать PEM")
	}

	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	pub, ok := pubInterface.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("ключ не является RSA-публичным")
	}

	return pub, nil
}

func HybridEncrypt(pub *rsa.PublicKey, plaintext []byte) (encryptedKey, ciphertext []byte, err error) {
	// Генерируем AES-ключ (256 бит)
	aesKey := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, aesKey); err != nil {
		return nil, nil, err
	}

	// Шифруем данные с помощью AES-GCM
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, err
	}

	ciphertext = gcm.Seal(nonce, nonce, plaintext, nil)

	// Шифруем AES-ключ с помощью RSA
	encryptedKey, err = rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		pub,
		aesKey,
		nil,
	)
	if err != nil {
		return nil, nil, err
	}

	return encryptedKey, ciphertext, nil
}

func ReadPrivateKey(filename string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, fmt.Errorf("не удалось декодировать PEM")
	}

	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		if rsaKey, ok := key.(*rsa.PrivateKey); ok {
			return rsaKey, nil
		}
		return nil, fmt.Errorf("ключ не является RSA")
	}

	return priv, nil
}

func HybridDecrypt(priv *rsa.PrivateKey, encryptedKey, ciphertext []byte) ([]byte, error) {
	aesKey, err := rsa.DecryptOAEP(
		sha256.New(),
		rand.Reader,
		priv,
		encryptedKey,
		nil,
	)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext слишком короткий")
	}

	nonce, encrypted := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, encrypted, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
