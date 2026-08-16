//go:build !linux && !darwin && !windows

package config

import "errors"

func platformFingerprint() ([]byte, error) {
	return nil, errors.New("machine fingerprint not supported on this platform")
}
