//go:build darwin

package install

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/achar-pranav/captive-bypass/internal/config"
)

const (
	serveLabel = "com.user.captive-bypass.serve"
	watchLabel = "com.user.captive-bypass.watch"
)

const servePlistTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.user.captive-bypass.serve</string>
    <key>ProgramArguments</key>
    <array>
        <string>%s</string>
        <string>serve</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardErrorPath</key>
    <string>%s/serve.err.log</string>
    <key>StandardOutPath</key>
    <string>%s/serve.out.log</string>
</dict>
</plist>`

const watchPlistTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.user.captive-bypass.watch</string>
    <key>ProgramArguments</key>
    <array>
        <string>%s</string>
        <string>watch</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardErrorPath</key>
    <string>%s/watch.err.log</string>
    <key>StandardOutPath</key>
    <string>%s/watch.out.log</string>
</dict>
</plist>`

func agentDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Library", "LaunchAgents"), nil
}

func plistPath(label string) (string, error) {
	dir, err := agentDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, label+".plist"), nil
}

func Enable() error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve executable: %w", err)
	}

	dir, err := agentDir()
	if err != nil {
		return fmt.Errorf("resolve LaunchAgents dir: %w", err)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating LaunchAgents dir: %w", err)
	}

	logDir := config.DefaultDir()
	_ = os.MkdirAll(logDir, 0o700)

	servePath, _ := plistPath(serveLabel)
	watchPath, _ := plistPath(watchLabel)

	if err := os.WriteFile(servePath, []byte(fmt.Sprintf(servePlistTemplate, exe, logDir, logDir)), 0o644); err != nil {
		return fmt.Errorf("write serve plist: %w", err)
	}
	if err := os.WriteFile(watchPath, []byte(fmt.Sprintf(watchPlistTemplate, exe, logDir, logDir)), 0o644); err != nil {
		return fmt.Errorf("write watch plist: %w", err)
	}

	_ = exec.Command("launchctl", "unload", servePath).Run()
	_ = exec.Command("launchctl", "unload", watchPath).Run()

	if out, err := exec.Command("launchctl", "load", "-w", servePath).CombinedOutput(); err != nil {
		return fmt.Errorf("launchctl load serve: %w: %s", err, out)
	}
	if out, err := exec.Command("launchctl", "load", "-w", watchPath).CombinedOutput(); err != nil {
		return fmt.Errorf("launchctl load watch: %w: %s", err, out)
	}

	return nil
}

func Disable() error {
	var firstErr error
	for _, label := range []string{serveLabel, watchLabel} {
		p, err := plistPath(label)
		if err != nil {
			continue
		}
		if out, err := exec.Command("launchctl", "unload", "-w", p).CombinedOutput(); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("launchctl unload %s: %w: %s", label, err, out)
		}
		_ = os.Remove(p)
	}
	return firstErr
}

func Status() (bool, error) {
	out, err := exec.Command("launchctl", "list", serveLabel).CombinedOutput()
	if err != nil {
		return false, nil
	}
	return len(out) > 0, nil
}

func Uninstall() error {
	if err := Disable(); err != nil {
		fmt.Fprintln(os.Stderr, "install:", err)
	}
	return os.RemoveAll(config.DefaultDir())
}
