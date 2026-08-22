//go:build windows

package windows

import (
	"fmt"
	"os/exec"

	"github.com/achar-pranav/captive-bypass/backends"
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

func (b *Backend) Scan() ([]backends.AP, error) {
	out, err := exec.Command("netsh", "wlan", "show", "networks", "mode=bssid").Output()
	if err != nil {
		return nil, fmt.Errorf("netsh wlan show networks: %w", err)
	}
	return parseNetworks(string(out)), nil
}
