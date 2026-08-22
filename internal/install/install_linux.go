//go:build linux

package install

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/achar-pranav/captive-bypass/internal/config"
)

const unitName = "captive-bypass.service"

const unitTemplate = `[Unit]
Description=captive-bypass captive portal watcher
After=network-online.target

[Service]
ExecStart=%s serve
Restart=on-failure
RestartSec=3

[Install]
WantedBy=default.target
`

func Enable() error {
	if _, err := exec.LookPath("iw"); err != nil {
		return errors.New("iw not found — install the iw package (e.g. 'sudo apt install iw'); captive-bypass needs it to read WiFi state")
	}
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve executable: %w", err)
	}
	if err := writeUnit(exe); err != nil {
		return err
	}
	if out, err := exec.Command("systemctl", "--user", "daemon-reload").CombinedOutput(); err != nil {
		return fmt.Errorf("daemon-reload: %w: %s", err, out)
	}
	if out, err := exec.Command("systemctl", "--user", "enable", "--now", unitName).CombinedOutput(); err != nil {
		return fmt.Errorf("enable service: %w: %s", err, out)
	}
	return nil
}

func Disable() error {
	var firstErr error
	out, err := exec.Command("systemctl", "--user", "disable", "--now", unitName).CombinedOutput()
	if err != nil && firstErr == nil {
		firstErr = fmt.Errorf("disable service: %w: %s", err, out)
	}
	os.Remove(unitFilePath())
	if out, err := exec.Command("systemctl", "--user", "daemon-reload").CombinedOutput(); err != nil && firstErr == nil {
		firstErr = fmt.Errorf("daemon-reload: %w: %s", err, out)
	}
	return firstErr
}

func Status() (bool, error) {
	out, err := exec.Command("systemctl", "--user", "is-enabled", unitName).Output()
	if err != nil {
		return false, nil
	}
	return string(out) == "enabled\n", nil
}

func Uninstall() error {
	if err := Disable(); err != nil {
		fmt.Fprintln(os.Stderr, "install:", err)
	}
	return os.RemoveAll(config.DefaultDir())
}

func unitFilePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".config", "systemd", "user", unitName)
	}
	return filepath.Join(home, ".config", "systemd", "user", unitName)
}

func writeUnit(exe string) error {
	p := unitFilePath()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	return os.WriteFile(p, []byte(fmt.Sprintf(unitTemplate, exe)), 0o644)
}
