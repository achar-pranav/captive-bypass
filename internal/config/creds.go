package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"

	"golang.org/x/crypto/scrypt"
)

const (
	keyLen   = 32
	scryptN  = 1 << 15
	scryptR  = 8
	scryptP  = 1
	saltLen  = 16
	nonceLen = 12
)

var ErrNoCreds = errors.New("no credentials stored")

type credsBlob struct {
	Username   string `json:"username"`
	Salt       []byte `json:"salt"`
	Nonce      []byte `json:"nonce"`
	Ciphertext []byte `json:"ciphertext"`
}

func (c *Config) SetCreds(fp []byte, username, password string) error {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return err
	}
	key, err := deriveKey(fp, salt)
	if err != nil {
		return err
	}
	nonce := make([]byte, nonceLen)
	if _, err := rand.Read(nonce); err != nil {
		return err
	}
	aead, err := newGCM(key)
	if err != nil {
		return err
	}
	c.Creds = credsBlob{
		Username:   username,
		Salt:       salt,
		Nonce:      nonce,
		Ciphertext: aead.Seal(nil, nonce, []byte(password), nil),
	}
	return nil
}

func (c *Config) GetCreds(fp []byte) (string, string, error) {
	if len(c.Creds.Ciphertext) == 0 {
		return "", "", ErrNoCreds
	}
	key, err := deriveKey(fp, c.Creds.Salt)
	if err != nil {
		return "", "", err
	}
	aead, err := newGCM(key)
	if err != nil {
		return "", "", err
	}
	pt, err := aead.Open(nil, c.Creds.Nonce, c.Creds.Ciphertext, nil)
	if err != nil {
		return "", "", err
	}
	return c.Creds.Username, string(pt), nil
}

func deriveKey(fp, salt []byte) ([]byte, error) {
	return scrypt.Key(fp, salt, scryptN, scryptR, scryptP, keyLen)
}

func newGCM(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}
