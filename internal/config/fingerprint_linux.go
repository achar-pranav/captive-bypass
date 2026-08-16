//go:build linux

package config

import (
	"os"
	"strings"
)

func platformFingerprint() ([]byte, error) {
	b, err := os.ReadFile("/etc/machine-id")
	if err != nil {
		return nil, err
	}
	return []byte(strings.TrimSpace(string(b))), nil
}
