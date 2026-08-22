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

var (
	ErrNoCreds     = errors.New("no credentials stored")
	ErrNoActiveSet = errors.New("no active credential set")
	ErrUnknownSet  = errors.New("unknown credential set")
)

type CredSet struct {
	Name       string `json:"name"`
	Username   string `json:"username"`
	Salt       []byte `json:"salt"`
	Nonce      []byte `json:"nonce"`
	Ciphertext []byte `json:"ciphertext"`
}

type credsBlob struct {
	Username   string `json:"username"`
	Salt       []byte `json:"salt"`
	Nonce      []byte `json:"nonce"`
	Ciphertext []byte `json:"ciphertext"`
}

func (c *Config) SetCredSet(fp []byte, name, username, password string) error {
	if name == "" {
		name = "default"
	}
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return err
	}
	nonce := make([]byte, nonceLen)
	if _, err := rand.Read(nonce); err != nil {
		return err
	}
	key, err := deriveKey(fp, salt)
	if err != nil {
		return err
	}
	aead, err := newGCM(key)
	if err != nil {
		return err
	}
	cs := CredSet{
		Name:       name,
		Username:   username,
		Salt:       salt,
		Nonce:      nonce,
		Ciphertext: aead.Seal(nil, nonce, []byte(password), nil),
	}
	for i := range c.CredSets {
		if c.CredSets[i].Name == name {
			c.CredSets[i] = cs
			return nil
		}
	}
	c.CredSets = append(c.CredSets, cs)
	if c.ActiveSet == "" {
		c.ActiveSet = name
	}
	return nil
}

func (c *Config) DeleteCredSet(name string) error {
	for i := range c.CredSets {
		if c.CredSets[i].Name == name {
			c.CredSets = append(c.CredSets[:i], c.CredSets[i+1:]...)
			if c.ActiveSet == name {
				c.ActiveSet = ""
			}
			return nil
		}
	}
	return ErrUnknownSet
}

func (c *Config) SetActiveSet(name string) error {
	for i := range c.CredSets {
		if c.CredSets[i].Name == name {
			c.ActiveSet = name
			return nil
		}
	}
	return ErrUnknownSet
}

func (c *Config) ActiveUser() (string, error) {
	cs := c.findActive()
	if cs == nil {
		return "", ErrNoActiveSet
	}
	return cs.Username, nil
}

func (c *Config) GetActiveCreds(fp []byte) (string, string, error) {
	cs := c.findActive()
	if cs == nil {
		return "", "", ErrNoCreds
	}
	key, err := deriveKey(fp, cs.Salt)
	if err != nil {
		return "", "", err
	}
	aead, err := newGCM(key)
	if err != nil {
		return "", "", err
	}
	pt, err := aead.Open(nil, cs.Nonce, cs.Ciphertext, nil)
	if err != nil {
		return "", "", err
	}
	return cs.Username, string(pt), nil
}

func (c *Config) findActive() *CredSet {
	if c.ActiveSet != "" {
		for i := range c.CredSets {
			if c.CredSets[i].Name == c.ActiveSet {
				return &c.CredSets[i]
			}
		}
	}
	return nil
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
