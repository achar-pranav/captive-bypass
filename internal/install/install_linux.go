//go:build linux

package install

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/achar-pranav/captive-bypass/internal/config"
)

const (
	unitName  = "captive-bypass.service"
	hookPath  = "/etc/NetworkManager/dispatcher.d/90-captive-bypass"
	hookPerms = 0o755
)

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

const hookTemplate = `#!/bin/sh
SOCK=%q
case "$2" in
	up)
		printf 'connect-current\n' | curl -s --unix-socket "$SOCK" http://localhost/hook >/dev/null 2>&1 &
		;;
	pre-down)
		printf 'disconnect\n' | curl -s --unix-socket "$SOCK" http://localhost/hook >/dev/null 2>&1 &
		;;
esac
exit 0
`

func Enable() error {
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
	return installHook(hookContent(config.DefaultDir()))
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
	if err := removeHook(); err != nil && firstErr == nil {
		firstErr = err
	}
	return firstErr
}

func Status() (bool, error) {
	out, err := exec.Command("systemctl", "--user", "is-enabled", unitName).Output()
	if err != nil {
		return false, nil
	}
	enabled := string(out) == "enabled\n"
	hookInstalled, err := hookInstalled()
	if err != nil {
		return enabled, err
	}
	return enabled && hookInstalled, nil
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

func hookContent(sockDir string) []byte {
	return []byte(fmt.Sprintf(hookTemplate, filepath.Join(sockDir, "serve.sock")))
}

func installHook(content []byte) error {
	tmp, err := os.CreateTemp("", "captive-bypass-hook-*")
	if err != nil {
		return err
	}
	path := tmp.Name()
	defer os.Remove(path)
	if _, err := tmp.Write(content); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(path, hookPerms); err != nil {
		return err
	}
	out, err := exec.Command("pkexec", "install", "-D", "-m", "0755", path, hookPath).CombinedOutput()
	if err != nil {
		return fmt.Errorf("pkexec hook install: %w: %s", err, out)
	}
	return nil
}

func removeHook() error {
	out, err := exec.Command("pkexec", "rm", "-f", hookPath).CombinedOutput()
	if err != nil {
		return fmt.Errorf("pkexec hook removal: %w: %s", err, out)
	}
	return nil
}

func hookInstalled() (bool, error) {
	fi, err := os.Stat(hookPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	content, err := os.ReadFile(hookPath)
	if err != nil {
		return false, err
	}
	want := hookContent(config.DefaultDir())
	return fi.Mode().Perm() == hookPerms && string(content) == string(want), nil
}
