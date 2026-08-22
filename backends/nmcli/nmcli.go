package nmcli

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/achar-pranav/captive-bypass/backends"
)

var ErrNoSignal = errors.New("no signal reading")

type Backend struct{}

func New() *Backend { return &Backend{} }

var _ backends.Backend = (*Backend)(nil)

func (b *Backend) ActiveSSID() (string, error) {
	out, err := run("nmcli", "-t", "-f", "active,ssid", "dev", "wifi")
	if err != nil {
		return "", err
	}
	return parseActive(out), nil
}

func (b *Backend) ActiveBSSID() (string, error) {
	out, err := run("nmcli", "-t", "-f", "active,bssid", "dev", "wifi")
	if err != nil {
		return "", err
	}
	return parseActive(out), nil
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
	out, err := run("nmcli", "networking", "connectivity")
	if err != nil {
		return false, err
	}
	return parseConnectivity(out), nil
}

func (b *Backend) iface() (string, error) {
	out, err := run("nmcli", "-t", "-f", "NAME,DEVICE,TYPE", "connection", "show", "--active")
	if err != nil {
		return "", err
	}
	iface := parseIface(out)
	if iface == "" {
		return "", fmt.Errorf("no active wifi connection")
	}
	return iface, nil
}

func parseIface(out string) string {
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		rest, found := strings.CutSuffix(line, ":802-11-wireless")
		if !found {
			continue
		}
		if i := strings.LastIndex(rest, ":"); i >= 0 && rest[i+1:] != "" {
			return rest[i+1:]
		}
	}
	return ""
}

func parseConnectivity(out string) bool {
	switch strings.TrimSpace(out) {
	case "full", "limited", "portal":
		return true
	default:
		return false
	}
}

func run(name string, args ...string) (string, error) {
	out, err := exec.Command(name, args...).Output()
	if err != nil {
		return "", fmt.Errorf("%s %s: %w", name, strings.Join(args, " "), err)
	}
	return string(out), nil
}

func parseActive(out string) string {
	for _, line := range strings.Split(out, "\n") {
		if rest, found := strings.CutPrefix(line, "yes:"); found {
			return unescape(strings.TrimSuffix(rest, "\r"))
		}
	}
	return ""
}

func unescape(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			i++
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

func parseSignal(out string) (int, error) {
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "signal:") {
			continue
		}
		field := strings.Fields(strings.TrimPrefix(line, "signal:"))
		if len(field) < 1 {
			break
		}
		var dbm int
		if _, err := fmt.Sscanf(field[0], "%d", &dbm); err == nil {
			return dbm, nil
		}
		break
	}
	return 0, ErrNoSignal
}
