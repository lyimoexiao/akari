package yggdrasil

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

// KeyManager handles RSA key pair operations for Yggdrasil property signing.
type KeyManager struct {
	privateKey *rsa.PrivateKey
	publicPEM  string // cached PEM-encoded public key
}

// NewKeyManager loads an existing key pair from keyDir or generates a new one.
func NewKeyManager(keyDir string) (*KeyManager, error) {
	path := filepath.Join(keyDir, privateKeyFile)

	key, err := loadPrivateKey(path)
	if err != nil {
		key, err = generateAndSaveKey(path)
		if err != nil {
			return nil, fmt.Errorf("key manager init: %w", err)
		}
	}

	pubPEM, err := publicKeyPEM(key)
	if err != nil {
		return nil, fmt.Errorf("marshal public key: %w", err)
	}

	return &KeyManager{privateKey: key, publicPEM: pubPEM}, nil
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
	priv, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("key is not RSA")
	}
	return priv, nil
}

func generateAndSaveKey(path string) (*rsa.PrivateKey, error) {
	os.MkdirAll(filepath.Dir(path), 0o755)

	key, err := rsa.GenerateKey(rand.Reader, keyBits)
	if err != nil {
		return nil, fmt.Errorf("generate RSA key: %w", err)
	}

	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("marshal private key: %w", err)
	}

	block := &pem.Block{Type: "PRIVATE KEY", Bytes: der}
	data := pem.EncodeToMemory(block)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return nil, fmt.Errorf("write private key: %w", err)
	}

	return key, nil
}

func publicKeyPEM(key *rsa.PrivateKey) (string, error) {
	der, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		return "", err
	}
	block := &pem.Block{Type: "PUBLIC KEY", Bytes: der}
	return string(pem.EncodeToMemory(block)), nil
}

// SignSHA1 signs data with SHA1-RSA (PKCS #1 v1.5).
// Returns Base64-encoded signature.
func (km *KeyManager) SignSHA1(data []byte) (string, error) {
	hashed := sha1.Sum(data)
	sig, err := rsa.SignPKCS1v15(rand.Reader, km.privateKey, crypto.SHA1, hashed[:])
	if err != nil {
		return "", fmt.Errorf("rsa sign: %w", err)
	}
	return base64.StdEncoding.EncodeToString(sig), nil
}

// PublicKeyPEM returns the cached PEM-encoded public key.
func (km *KeyManager) PublicKeyPEM() string {
	return km.publicPEM
}
