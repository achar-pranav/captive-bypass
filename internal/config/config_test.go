package config

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

const (
	fpA = "fake-machine-fingerprint-A"
	fpB = "fake-machine-fingerprint-B"
)

func TestRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	cfg := Default()
	cfg.SSIDs = []string{"ELEMENT BLOCK", "ELEMENT-BLOCK-5G"}
	cfg.Paused = true
	if err := cfg.SetCreds([]byte(fpA), "1BI22CS123", "hunter2"); err != nil {
		t.Fatalf("SetCreds: %v", err)
	}
	if err := Save(path, cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	want := cfg
	want.Creds.Username = cfg.Creds.Username
	if !reflect.DeepEqual(got.SSIDs, want.SSIDs) {
		t.Errorf("SSIDs = %v, want %v", got.SSIDs, want.SSIDs)
	}
	if got.Paused != want.Paused {
		t.Errorf("Paused = %v, want %v", got.Paused, want.Paused)
	}
	if got.Portal != want.Portal {
		t.Errorf("Portal = %q, want %q", got.Portal, want.Portal)
	}
	if !reflect.DeepEqual(got.Timings, want.Timings) {
		t.Errorf("Timings = %+v, want %+v", got.Timings, want.Timings)
	}

	user, pass, err := got.GetCreds([]byte(fpA))
	if err != nil {
		t.Fatalf("Creds: %v", err)
	}
	if user != "1BI22CS123" {
		t.Errorf("username = %q, want %q", user, "1BI22CS123")
	}
	if pass != "hunter2" {
		t.Errorf("password = %q, want %q", pass, "hunter2")
	}
}

func TestNoPlaintextPasswordOnDisk(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	cfg := Default()
	if err := cfg.SetCreds([]byte(fpA), "1BI22CS123", "hunter2"); err != nil {
		t.Fatalf("SetCreds: %v", err)
	}
	if err := Save(path, cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if strings.Contains(string(b), "hunter2") {
		t.Error("plaintext password found in config.json")
	}
}

func TestWrongFingerprintFails(t *testing.T) {
	cfg := Default()
	if err := cfg.SetCreds([]byte(fpA), "1BI22CS123", "hunter2"); err != nil {
		t.Fatalf("SetCreds: %v", err)
	}
	if _, _, err := cfg.GetCreds([]byte(fpB)); err == nil {
		t.Error("Creds with a different fingerprint succeeded, want error")
	}
}

func TestCredsNotSet(t *testing.T) {
	cfg := Default()
	if _, _, err := cfg.GetCreds([]byte(fpA)); err != ErrNoCreds {
		t.Errorf("Creds = %v, want ErrNoCreds", err)
	}
}

func TestFileAndDirPerms(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission bits not enforced the same way on windows")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	cfg := Default()
	if err := cfg.SetCreds([]byte(fpA), "1BI22CS123", "hunter2"); err != nil {
		t.Fatalf("SetCreds: %v", err)
	}
	if err := Save(path, cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}
	fi, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if fi.Mode().Perm() != 0o600 {
		t.Errorf("file mode = %o, want 600", fi.Mode().Perm())
	}
	di, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("Stat dir: %v", err)
	}
	if di.Mode().Perm() != 0o700 {
		t.Errorf("dir mode = %o, want 700", di.Mode().Perm())
	}
}

func TestMissingConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nope.json")
	_, err := Load(path)
	if err != ErrNoConfig {
		t.Errorf("Load = %v, want ErrNoConfig", err)
	}
}

func TestDefaults(t *testing.T) {
	t.Setenv("CAPTIVE_BYPASS_PORTAL", "http://127.0.0.1:9999")
	t.Setenv("CAPTIVE_BYPASS_RETRY_DELAY", "2")
	cfg := Default()
	if cfg.Portal != "http://127.0.0.1:9999" {
		t.Errorf("Portal = %q, want env override", cfg.Portal)
	}
	if cfg.Timings.RetryDelay != 2 {
		t.Errorf("RetryDelay = %d, want 2", cfg.Timings.RetryDelay)
	}
}
