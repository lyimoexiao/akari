package yggdrasiladapter

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
)

const (
	privateKeyFile = "yggdrasil_private_key.pem"
	keyBits        = 2048
)

type KeyManager struct {
	privateKey *rsa.PrivateKey
	publicPEM  string
}

func NewKeyManager(keyDir string) (*KeyManager, error) {
	path := filepath.Join(keyDir, privateKeyFile)
	key, err := loadPrivateKey(path)
	if err != nil {
		key, err = generateAndSaveKey(path)
		if err != nil {
			return nil, fmt.Errorf("key manager init: %w", err)
		}
	}
	publicPEM, err := publicKeyPEM(key)
	if err != nil {
		return nil, fmt.Errorf("marshal public key: %w", err)
	}
	return &KeyManager{privateKey: key, publicPEM: publicPEM}, nil
}

func (m *KeyManager) Sign(data []byte) (string, error) {
	hashed := sha1.Sum(data)
	signature, err := rsa.SignPKCS1v15(rand.Reader, m.privateKey, crypto.SHA1, hashed[:])
	if err != nil {
		return "", fmt.Errorf("rsa sign: %w", err)
	}
	return base64.StdEncoding.EncodeToString(signature), nil
}

func (m *KeyManager) PublicKey() string {
	return m.publicPEM
}

func loadPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("no PEM block in %s", path)
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}
	privateKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("key is not RSA")
	}
	return privateKey, nil
}

func generateAndSaveKey(path string) (*rsa.PrivateKey, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create key directory: %w", err)
	}
	key, err := rsa.GenerateKey(rand.Reader, keyBits)
	if err != nil {
		return nil, fmt.Errorf("generate RSA key: %w", err)
	}
	encoded, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("marshal private key: %w", err)
	}
	data := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: encoded})
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return nil, fmt.Errorf("write private key: %w", err)
	}
	return key, nil
}

func publicKeyPEM(key *rsa.PrivateKey) (string, error) {
	encoded, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		return "", err
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: encoded})), nil
}
