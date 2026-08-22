//go:build windows

package install

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/achar-pranav/captive-bypass/internal/config"
)

const (
	serveTask = "captive-bypass serve"
	watchTask = "captive-bypass watch"
)

func Enable() error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve executable: %w", err)
	}
	if err := createTask(serveTask, exe+" serve"); err != nil {
		return err
	}
	return createTask(watchTask, exe+" watch")
}

func Disable() error {
	var firstErr error
	for _, task := range []string{serveTask, watchTask} {
		if err := deleteTask(task); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func Status() (bool, error) {
	out, err := exec.Command("schtasks", "/Query", "/TN", serveTask).CombinedOutput()
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

func q(s string) string { return "\"" + s + "\"" }

func createTask(name, action string) error {
	out, err := exec.Command("schtasks", "/Create", "/F", "/SC", "ONLOGON", "/RL", "LIMITED",
		"/TN", name, "/TR", q(action)).CombinedOutput()
	if err != nil {
		return fmt.Errorf("schtasks create %q: %w: %s", name, err, out)
	}
	return nil
}

func deleteTask(name string) error {
	out, err := exec.Command("schtasks", "/Delete", "/F", "/TN", name).CombinedOutput()
	if err != nil {
		return fmt.Errorf("schtasks delete %q: %w: %s", name, err, out)
	}
	return nil
}
