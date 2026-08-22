//go:build linux

package install

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

func TestHookContent(t *testing.T) {
	dir := t.TempDir()
	content := string(hookContent(dir))
	sock := filepath.Join(dir, "serve.sock")
	if !strings.Contains(content, sock) {
		t.Errorf("hook missing socket path %s:\n%s", sock, content)
	}
	for _, want := range []string{"connect-current", "disconnect", "up)", "pre-down)", "#!/bin/sh"} {
		if !strings.Contains(content, want) {
			t.Errorf("hook missing %q:\n%s", want, content)
		}
	}
}

func TestUnitContent(t *testing.T) {
	body := fmt.Sprintf(unitTemplate, "/opt/cb/captive-bypass")
	if !strings.Contains(body, "ExecStart=/opt/cb/captive-bypass serve") {
		t.Errorf("unit missing ExecStart:\n%s", body)
	}
	if !strings.Contains(body, "WantedBy=default.target") {
		t.Errorf("unit missing install target:\n%s", body)
	}
}
