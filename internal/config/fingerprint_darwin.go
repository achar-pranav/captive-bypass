//go:build darwin

package config

import "syscall"

func platformFingerprint() ([]byte, error) {
	s, err := syscall.Sysctl("kern.uuid")
	if err != nil {
		return nil, err
	}
	return []byte(s), nil
}
