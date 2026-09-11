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

func (c *Config) SetCredSet(fp []byte, username, password string) error {
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
		Username:   username,
		Salt:       salt,
		Nonce:      nonce,
		Ciphertext: aead.Seal(nil, nonce, []byte(password), nil),
	}
	for i := range c.CredSets {
		if c.CredSets[i].Username == username {
			c.CredSets[i] = cs
			return nil
		}
	}
	c.CredSets = append(c.CredSets, cs)
	// We no longer auto-set ActiveSet here. We wait for explicit selection.
	return nil
}

func (c *Config) DeleteCredSet(username string) error {
	for i := range c.CredSets {
		if c.CredSets[i].Username == username {
			c.CredSets = append(c.CredSets[:i], c.CredSets[i+1:]...)
			if c.ActiveSet == username {
				c.ActiveSet = ""
			}
			return nil
		}
	}
	return ErrUnknownSet
}

func (c *Config) SetActiveSet(username string) error {
	for i := range c.CredSets {
		if c.CredSets[i].Username == username {
			c.ActiveSet = username
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

func (c *Config) GetCredsByUsername(fp []byte, username string) (string, string, error) {
	for i := range c.CredSets {
		if c.CredSets[i].Username == username {
			key, err := deriveKey(fp, c.CredSets[i].Salt)
		if err != nil {
			return "", "", err
		}
		aead, err := newGCM(key)
		if err != nil {
			return "", "", err
		}
		pt, err := aead.Open(nil, c.CredSets[i].Nonce, c.CredSets[i].Ciphertext, nil)
		if err != nil {
			return "", "", err
		}
		return c.CredSets[i].Username, string(pt), nil
		}
	}
	return "", "", ErrUnknownSet
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
			if c.CredSets[i].Username == c.ActiveSet {
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
