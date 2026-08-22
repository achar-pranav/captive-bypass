//go:build windows

package windows

import (
	"fmt"
	"os/exec"
)

type Backend struct{}

func New() *Backend { return &Backend{} }

func (b *Backend) snapshot() (snapshot, error) {
	out, err := exec.Command("netsh", "wlan", "show", "interfaces").Output()
	if err != nil {
		return snapshot{}, fmt.Errorf("netsh wlan show interfaces: %w", err)
	}
	return parseInterfaces(string(out)), nil
}

func (b *Backend) ActiveSSID() (string, error) {
	s, err := b.snapshot()
	return s.SSID, err
}

func (b *Backend) ActiveBSSID() (string, error) {
	s, err := b.snapshot()
	return s.BSSID, err
}

func (b *Backend) Signal() (int, error) {
	s, err := b.snapshot()
	return s.Signal, err
}

func (b *Backend) Up() (bool, error) {
	s, err := b.snapshot()
	return s.Up, err
}
