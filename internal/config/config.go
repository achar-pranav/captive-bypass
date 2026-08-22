package config

import (
	"encoding/json"
	"errors"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
)

const DefaultPortal = "https://rr.pes.edu:8090"

var ErrNoConfig = errors.New("no config file")

type Config struct {
	SSIDs     []string  `json:"ssids"`
	Paused    bool      `json:"paused"`
	Portal    string    `json:"portal"`
	Timings   Timings   `json:"timings"`
	CredSets  []CredSet `json:"cred_sets"`
	ActiveSet string    `json:"active_set"`
	Vanguard  bool      `json:"vanguard"`
	Creds     credsBlob `json:"creds,omitempty"`
}

type Timings struct {
	RetryDelay     int `json:"retry_delay"`
	LoginCooldown  int `json:"login_cooldown"`
	LogoutCooldown int `json:"logout_cooldown"`
}

func Default() *Config {
	return &Config{
		Portal: envOr("CAPTIVE_BYPASS_PORTAL", DefaultPortal),
		Timings: Timings{
			RetryDelay:     intEnv("CAPTIVE_BYPASS_RETRY_DELAY", 5),
			LoginCooldown:  intEnv("CAPTIVE_BYPASS_LOGIN_COOLDOWN", 60),
			LogoutCooldown: intEnv("CAPTIVE_BYPASS_LOGOUT_COOLDOWN", 10),
		},
	}
}

func DefaultDir() string {
	if d := os.Getenv("CAPTIVE_BYPASS_CONFIG"); d != "" {
		return d
	}
	if h := os.Getenv("HOME"); h != "" {
		return filepath.Join(h, ".config", "captive-bypass")
	}
	if u, err := user.Current(); err == nil {
		return filepath.Join(u.HomeDir, ".config", "captive-bypass")
	}
	return filepath.Join(".", ".config", "captive-bypass")
}

func Load(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Default(), ErrNoConfig
		}
		return nil, err
	}
	c := Default()
	if err := json.Unmarshal(b, c); err != nil {
		return nil, err
	}
	c.migrateLegacyCreds()
	return c, nil
}

func (c *Config) migrateLegacyCreds() {
	if len(c.CredSets) > 0 || len(c.Creds.Ciphertext) == 0 {
		return
	}
	c.CredSets = append(c.CredSets, CredSet{
		Name:       "default",
		Username:   c.Creds.Username,
		Salt:       c.Creds.Salt,
		Nonce:      c.Creds.Nonce,
		Ciphertext: c.Creds.Ciphertext,
	})
	c.ActiveSet = "default"
}

func Save(path string, c *Config) error {
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	if err := os.Chmod(dir, 0o700); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	if err := os.Chmod(tmp, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func envOr(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func intEnv(k string, d int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return d
}
