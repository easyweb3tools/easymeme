package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/ethereum/go-ethereum/crypto"
)

func EncryptPrivateKey(masterKey string, privateKey []byte) ([]byte, error) {
	if masterKey == "" {
		return nil, fmt.Errorf("WALLET_MASTER_KEY is required")
	}
	hash := sha256.Sum256([]byte(masterKey))
	block, err := aes.NewCipher(hash[:])
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	ciphertext := gcm.Seal(nil, nonce, privateKey, nil)
	out := append(nonce, ciphertext...)
	return []byte(hex.EncodeToString(out)), nil
}

func DecryptPrivateKey(masterKey string, cipherHex []byte) (*ecdsa.PrivateKey, error) {
	if masterKey == "" {
		return nil, fmt.Errorf("WALLET_MASTER_KEY is required")
	}
	raw, err := hex.DecodeString(string(cipherHex))
	if err != nil {
		return nil, err
	}
	hash := sha256.Sum256([]byte(masterKey))
	block, err := aes.NewCipher(hash[:])
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if len(raw) < gcm.NonceSize() {
		return nil, fmt.Errorf("invalid ciphertext")
	}
	nonce := raw[:gcm.NonceSize()]
	ciphertext := raw[gcm.NonceSize():]
	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}
	return crypto.ToECDSA(plain)
}
