package config

import (
	"os"
	"path/filepath"
	"testing"
)

func testFP(t *testing.T) []byte {
	t.Helper()
	return []byte("test-fingerprint-0123456789abcdef")
}

func TestCredSetRoundtrip(t *testing.T) {
	c := Default()
	if err := c.SetCredSet(testFP(t), "work", "PES123", "secret"); err != nil {
		t.Fatal(err)
	}
	if c.ActiveSet != "work" {
		t.Fatalf("active = %q, want work", c.ActiveSet)
	}
	user, pass, err := c.GetActiveCreds(testFP(t))
	if err != nil || user != "PES123" || pass != "secret" {
		t.Fatalf("got %q/%q, %v", user, pass, err)
	}
	if _, _, err := c.GetActiveCreds([]byte("wrong-fp")); err == nil {
		t.Fatal("expected decrypt failure with wrong fingerprint")
	}
}

func TestCredSetUpsertAndSwitch(t *testing.T) {
	c := Default()
	fp := testFP(t)
	if err := c.SetCredSet(fp, "a", "u1", "p1"); err != nil {
		t.Fatal(err)
	}
	if err := c.SetCredSet(fp, "b", "u2", "p2"); err != nil {
		t.Fatal(err)
	}
	if len(c.CredSets) != 2 {
		t.Fatalf("sets = %d, want 2", len(c.CredSets))
	}
	if err := c.SetActiveSet("b"); err != nil {
		t.Fatal(err)
	}
	user, pass, _ := c.GetActiveCreds(fp)
	if user != "u2" || pass != "p2" {
		t.Fatalf("active resolved to %q/%q", user, pass)
	}
	if err := c.SetCredSet(fp, "b", "u3", "p3"); err != nil {
		t.Fatal(err)
	}
	if got := len(c.CredSets); got != 2 {
		t.Fatalf("upsert created set #%d", got)
	}
	user, pass, _ = c.GetActiveCreds(fp)
	if user != "u3" || pass != "p3" {
		t.Fatalf("upsert not applied: %q/%q", user, pass)
	}
}

func TestDeleteCredSetClearsActive(t *testing.T) {
	c := Default()
	fp := testFP(t)
	c.SetCredSet(fp, "only", "u", "p")
	if err := c.DeleteCredSet("only"); err != nil {
		t.Fatal(err)
	}
	if c.ActiveSet != "" || len(c.CredSets) != 0 {
		t.Fatalf("delete left %+v active=%q", c.CredSets, c.ActiveSet)
	}
	if _, _, err := c.GetActiveCreds(fp); err != ErrNoCreds {
		t.Fatalf("err = %v, want ErrNoCreds", err)
	}
	if err := c.DeleteCredSet("nope"); err != ErrUnknownSet {
		t.Fatalf("err = %v, want ErrUnknownSet", err)
	}
}

func TestLegacyMigration(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "config.json")
	old := `{"ssids":["X"],"creds":{"username":"OLD","salt":"c2FsdA==","nonce":"bm9uY2U=","ciphertext":"ZGF0YQ=="}}`
	if err := writeFile(p, old); err != nil {
		t.Fatal(err)
	}
	c, err := Load(p)
	if err != nil && err != ErrNoConfig {
		t.Fatal(err)
	}
	if c.ActiveSet != "default" || len(c.CredSets) != 1 || c.CredSets[0].Username != "OLD" {
		t.Fatalf("migration produced %+v active=%q", c.CredSets, c.ActiveSet)
	}
}

func writeFile(p, content string) error {
	return os.WriteFile(p, []byte(content), 0o600)
}
