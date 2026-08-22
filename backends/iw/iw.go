//go:build linux

// Package iw reads WiFi state straight from the kernel interface layer,
// below any manager (NetworkManager, iwd, wpa_supplicant). One reader for
// all of them.
package iw

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/achar-pranav/captive-bypass/backends"
	"github.com/achar-pranav/captive-bypass/backends/nmcli"
)

var ErrNoWifi = errors.New("no wireless interface found")

type Backend struct {
	nmcliPath string
}

func New() *Backend {
	p, err := exec.LookPath("nmcli")
	if err != nil {
		p = ""
	}
	return &Backend{nmcliPath: p}
}

func (b *Backend) iface() (string, error) {
	entries, err := os.ReadDir("/sys/class/net")
	if err != nil {
		return "", err
	}
	for _, e := range entries {
		if isWireless(filepath.Join("/sys/class/net", e.Name())) {
			return e.Name(), nil
		}
	}
	return "", ErrNoWifi
}

func isWireless(sysDir string) bool {
	fi, err := os.Stat(filepath.Join(sysDir, "phy80211"))
	return err == nil && fi.IsDir()
}

func (b *Backend) ActiveSSID() (string, error) { return b.linkField("ssid") }

func (b *Backend) ActiveBSSID() (string, error) {
	iface, err := b.iface()
	if err != nil {
		return "", err
	}
	out, err := run("iw", "dev", iface, "link")
	if err != nil {
		return "", err
	}
	if v := parseConnectedBSSID(out); v != "" {
		return v, nil
	}
	return "", errors.New("not associated")
}

func (b *Backend) Signal() (int, error) {
	iface, err := b.iface()
	if err != nil {
		return 0, err
	}
	out, err := run("iw", "dev", iface, "link")
	if err != nil {
		return 0, err
	}
	return parseSignal(out)
}

func (b *Backend) Up() (bool, error) {
	iface, err := b.iface()
	if err != nil {
		return false, err
	}
	carrier, err := os.ReadFile(filepath.Join("/sys/class/net", iface, "carrier"))
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(string(carrier)) == "1", nil
}

func (b *Backend) Scan() ([]backends.AP, error) {
	if b.nmcliPath == "" {
		return nil, errors.New("network scan unavailable (nmcli not installed)")
	}
	out, err := run(b.nmcliPath, "-t", "-f", "ssid,bssid,signal,security", "device", "wifi", "list")
	if err != nil {
		return nil, err
	}
	return nmcli.ParseScan(out), nil
}

func (b *Backend) linkField(field string) (string, error) {
	iface, err := b.iface()
	if err != nil {
		return "", err
	}
	out, err := run("iw", "dev", iface, "link")
	if err != nil {
		return "", err
	}
	v := parseLinkField(out, field)
	if v == "" {
		return "", fmt.Errorf("not associated")
	}
	return v, nil
}

func run(name string, args ...string) (string, error) {
	out, err := exec.Command(name, args...).Output()
	if err != nil {
		return "", fmt.Errorf("%s %s: %w", name, strings.Join(args, " "), err)
	}
	return string(out), nil
}

func parseLinkField(out, field string) string {
	for _, line := range strings.Split(out, "\n") {
		k, v, ok := strings.Cut(strings.TrimSpace(line), ":")
		if !ok || !strings.EqualFold(strings.TrimSpace(k), field) {
			continue
		}
		return strings.TrimSpace(v)
	}
	return ""
}

func parseSignal(out string) (int, error) {
	v := parseLinkField(out, "signal")
	if v == "" {
		return 0, errors.New("no signal reading")
	}
	var dbm int
	if _, err := fmt.Sscanf(v, "%d", &dbm); err != nil {
		return 0, errors.New("no signal reading")
	}
	return dbm, nil
}

func parseConnectedBSSID(out string) string {
	for _, line := range strings.Split(out, "\n") {
		if rest, ok := strings.CutPrefix(strings.TrimSpace(line), "Connected to "); ok {
			f := strings.Fields(rest)
			if len(f) > 0 && f[0] != "none" {
				return f[0]
			}
		}
	}
	return ""
}
