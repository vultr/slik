package munge

import (
	"crypto/rand"
	"crypto/sha256"
	b64 "encoding/base64"
	"encoding/hex"

	"golang.org/x/crypto/hkdf"
)

type MungeKey struct {
	Salt   []byte
	Info   []byte
	Secret []byte
}

// SecretRaw returns the secret as raw byte array
func (m *MungeKey) SecretRaw() []byte {
	return m.Secret
}

// SecretRawString returns the secret as string
func (m *MungeKey) SecretRawString() string {
	return string(m.Secret)
}

// SecretAsHex returns the secret hex encoded
func (m *MungeKey) SecretAsHex() string {
	return hex.EncodeToString(m.Secret)
}

// SecretBase64 returns the secret base64 encoded
func (m *MungeKey) SecretBase64() string {
	return b64.StdEncoding.EncodeToString(m.Secret)
}

// NewMungeKey creates a new munge key
func NewMungeKey() (*MungeKey, error) {
	// Underlying hash function for HMAC.
	hash := sha256.New

	// Cryptographically secure secret.
	secret := make([]byte, MungeKeyBytes)
	_, err := rand.Read(secret)
	if err != nil {
		return nil, err
	}

	// Non-secret salt, optional (can be nil).
	// Recommended: hash-length random value.
	salt := make([]byte, hash().Size())
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}

	// Non-secret context info, optional (can be nil).
	info := []byte(MungeKeyInfo)

	// Generate derived keys.
	hkdf.New(hash, secret, salt, info)

	return &MungeKey{
		Salt:   salt,
		Info:   info,
		Secret: secret,
	}, nil
}
