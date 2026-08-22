package serve

import (
	"bufio"
	"context"
	"github.com/achar-pranav/captive-bypass/backends"
	"net"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/achar-pranav/captive-bypass/internal/config"
	"github.com/achar-pranav/captive-bypass/internal/state"
)

type fakePortal struct {
	live      bool
	loginOK   bool
	loginMsg  string
	loginErr  error
	logoutErr error
	calls     []string
}

func (f *fakePortal) Login(ctx context.Context, u, p string) (bool, string, error) {
	f.calls = append(f.calls, "login")
	return f.loginOK, f.loginMsg, f.loginErr
}

func (f *fakePortal) Logout(ctx context.Context, u string) error {
	f.calls = append(f.calls, "logout")
	return f.logoutErr
}

func (f *fakePortal) Livecheck(ctx context.Context) (bool, error) {
	f.calls = append(f.calls, "livecheck")
	return f.live, nil
}

type fakeWifi struct{ up bool }

func (f *fakeWifi) ActiveSSID() (string, error)  { return "Campus", nil }
func (f *fakeWifi) ActiveBSSID() (string, error) { return "AA:BB:CC:DD:EE:FF", nil }
func (f *fakeWifi) Signal() (int, error)         { return -50, nil }
func (f *fakeWifi) Up() (bool, error)            { return f.up, nil }

func newTestServer(t *testing.T) (*Server, *fakePortal, *fakeWifi, *[]string) {
	t.Helper()
	dir := t.TempDir()
	fp, err := config.MachineFingerprint()
	if err != nil {
		t.Fatalf("fingerprint: %v", err)
	}
	cfg := config.Default()
	cfg.SSIDs = []string{"Campus"}
	if err := cfg.SetCredSet(fp, "default", "1BI22CS123", "pw"); err != nil {
		t.Fatalf("SetCreds: %v", err)
	}
	cfgPath := filepath.Join(dir, "config.json")
	if err := config.Save(cfgPath, cfg); err != nil {
		t.Fatalf("Save config: %v", err)
	}
	p := &fakePortal{}
	w := &fakeWifi{up: true}
	var notes []string
	s := &Server{
		ConfigPath:  cfgPath,
		StatePath:   filepath.Join(dir, "state"),
		SocketPath:  filepath.Join(dir, "serve.sock"),
		SettleDelay: 0,
		Portal:      p,
		Wifi:        w,
		Notify:      func(msg string) { notes = append(notes, msg) },
	}
	return s, p, w, &notes
}

func TestConnectLogsIn(t *testing.T) {
	s, p, _, notes := newTestServer(t)
	p.loginOK = true
	reply := s.Handle(context.Background(), "connect Campus")
	if reply != "ok logged-in" {
		t.Fatalf("reply = %q, want ok logged-in", reply)
	}
	for _, c := range p.calls {
		if c == "login" {
			goto found
		}
	}
	t.Fatalf("login not called: %v", p.calls)
found:
	st, err := state.Load(s.StatePath)
	if err != nil || st.Action != state.ActionLogin || st.BSSID != "AA:BB:CC:DD:EE:FF" {
		t.Errorf("state = %+v err %v, want login with BSSID", st, err)
	}
	if len(*notes) == 0 || !strings.Contains((*notes)[0], "Signed in on Campus") {
		t.Errorf("notes = %v, want signed-in notification", *notes)
	}
}

func TestConnectAlreadyOnlineSkipsLogin(t *testing.T) {
	s, p, _, notes := newTestServer(t)
	p.live = true
	reply := s.Handle(context.Background(), "connect Campus")
	if reply != "ok already-online" {
		t.Fatalf("reply = %q, want ok already-online", reply)
	}
	for _, c := range p.calls {
		if c == "login" {
			t.Fatalf("login called despite live check: %v", p.calls)
		}
	}
	if len(*notes) == 0 || !strings.Contains((*notes)[0], "Already online on Campus") {
		t.Errorf("notes = %v, want already-online notification", *notes)
	}
}

func TestConnectPausedIgnored(t *testing.T) {
	s, p, _, _ := newTestServer(t)
	cfg, _ := config.Load(s.ConfigPath)
	cfg.Paused = true
	config.Save(s.ConfigPath, cfg)
	reply := s.Handle(context.Background(), "connect Campus")
	if reply != "skip paused" {
		t.Fatalf("reply = %q, want skip paused", reply)
	}
	if len(p.calls) != 0 {
		t.Errorf("portal calls = %v, want none", p.calls)
	}
}

func TestConnectUnregisteredIgnored(t *testing.T) {
	s, p, _, _ := newTestServer(t)
	reply := s.Handle(context.Background(), "connect OtherNet")
	if reply != "skip unregistered" {
		t.Fatalf("reply = %q, want skip unregistered", reply)
	}
	if len(p.calls) != 0 {
		t.Errorf("portal calls = %v, want none", p.calls)
	}
}

func TestConnectLoginCooldown(t *testing.T) {
	s, p, _, _ := newTestServer(t)
	saveState(s.StatePath, state.ActionLogin, "")
	reply := s.Handle(context.Background(), "connect Campus")
	if reply != "skip login-cooldown" {
		t.Fatalf("reply = %q, want skip login-cooldown", reply)
	}
	for _, c := range p.calls {
		if c == "login" {
			t.Fatalf("login called during cooldown: %v", p.calls)
		}
	}
}

func TestConnectLimitMessage(t *testing.T) {
	s, p, _, notes := newTestServer(t)
	p.loginMsg = "You have reached maximum login limit"
	reply := s.Handle(context.Background(), "connect Campus")
	if reply != "ok login-limit" {
		t.Fatalf("reply = %q, want ok login-limit", reply)
	}
	if len(*notes) == 0 || !strings.HasPrefix((*notes)[0], "Login limit reached on Campus") {
		t.Errorf("notes = %v, want limit notification", *notes)
	}
}

func TestConnectRejected(t *testing.T) {
	s, p, _, _ := newTestServer(t)
	p.loginMsg = "authentication failure"
	reply := s.Handle(context.Background(), "connect Campus")
	if reply != "err rejected" {
		t.Fatalf("reply = %q, want err rejected", reply)
	}
	st, err := state.Load(s.StatePath)
	if err == nil && st.Action == state.ActionLogin {
		t.Errorf("state recorded login despite rejection")
	}
}

func TestDisconnectLogsOut(t *testing.T) {
	s, p, _, _ := newTestServer(t)
	reply := s.Handle(context.Background(), "disconnect")
	if reply != "ok logged-out" {
		t.Fatalf("reply = %q, want ok logged-out", reply)
	}
	for _, c := range p.calls {
		if c == "logout" {
			goto found
		}
	}
	t.Fatalf("logout not called: %v", p.calls)
found:
	st, err := state.Load(s.StatePath)
	if err != nil || st.Action != state.ActionLogout || st.BSSID != "" {
		t.Errorf("state = %+v err %v, want logout with empty BSSID", st, err)
	}
}

func TestShutdownLogoutWhenUp(t *testing.T) {
	s, p, w, _ := newTestServer(t)
	saveState(s.StatePath, state.ActionLogin, "")
	w.up = true
	s.shutdownLogout(context.Background())
	for _, c := range p.calls {
		if c == "logout" {
			return
		}
	}
	t.Fatalf("clean shutdown did not log out: %v", p.calls)
}

func TestShutdownNoLogoutWhenDown(t *testing.T) {
	s, p, w, _ := newTestServer(t)
	saveState(s.StatePath, state.ActionLogin, "")
	w.up = false
	s.shutdownLogout(context.Background())
	for _, c := range p.calls {
		if c == "logout" {
			t.Fatalf("logout attempted without network: %v", p.calls)
		}
	}
}

func TestConnectCurrentUsesBackend(t *testing.T) {
	s, p, _, _ := newTestServer(t)
	p.loginOK = true
	reply := s.Handle(context.Background(), "connect-current")
	if reply != "ok logged-in" {
		t.Fatalf("reply = %q, want ok logged-in", reply)
	}
}

func TestSocketRoundTrip(t *testing.T) {
	s, p, _, _ := newTestServer(t)
	p.loginOK = true
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- s.Run(ctx) }()
	time.Sleep(50 * time.Millisecond)

	conn, err := net.Dial("unix", s.SocketPath)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	if _, err := conn.Write([]byte("connect Campus\n")); err != nil {
		t.Fatalf("write: %v", err)
	}
	reply, err := bufio.NewReader(conn).ReadString('\n')
	conn.Close()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !strings.HasPrefix(reply, "ok logged-in") {
		t.Errorf("socket reply = %q, want ok logged-in", reply)
	}
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Run = %v, want nil", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not exit after cancel")
	}
}

func (f *fakeWifi) Scan() ([]backends.AP, error) { return nil, nil }
