package serve

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/achar-pranav/captive-bypass/backends"
	"github.com/achar-pranav/captive-bypass/backends/nmcli"
	"github.com/achar-pranav/captive-bypass/internal/config"
	"github.com/achar-pranav/captive-bypass/internal/portal"
	"github.com/achar-pranav/captive-bypass/internal/state"
)

const settleDelay = 1500 * time.Millisecond

type Portal interface {
	Login(ctx context.Context, username, password string) (bool, string, error)
	Logout(ctx context.Context, username string) error
	Livecheck(ctx context.Context) (bool, error)
}

type Server struct {
	ConfigPath  string
	StatePath   string
	SocketPath  string
	SettleDelay time.Duration
	Portal      Portal
	Wifi        backends.Backend
	Notify      func(msg string)

	mu     sync.Mutex
	cancel context.CancelFunc
}

func DefaultSocketPath() string {
	return filepath.Join(config.DefaultDir(), "serve.sock")
}

func New() *Server {
	dir := config.DefaultDir()
	return &Server{
		ConfigPath:  filepath.Join(dir, "config.json"),
		StatePath:   state.DefaultPath(),
		SocketPath:  DefaultSocketPath(),
		SettleDelay: settleDelay,
		Portal:      portal.New("", nil),
		Wifi:        nmcli.New(),
		Notify:      notifySend,
	}
}

var notifySend = func(msg string) {
	if err := exec.Command("notify-send", "captive-bypass", msg).Start(); err != nil {
		log.Printf("notify: %v", err)
	}
}

func (s *Server) Run(ctx context.Context) error {
	ctx, s.cancel = context.WithCancel(ctx)
	defer func() { s.cancel = nil }()

	if err := os.MkdirAll(filepath.Dir(s.SocketPath), 0o700); err != nil {
		return err
	}
	os.Remove(s.SocketPath)
	ln, err := net.Listen("unix", s.SocketPath)
	if err != nil {
		return fmt.Errorf("listen %s: %w", s.SocketPath, err)
	}
	defer os.Remove(s.SocketPath)

	go func() {
		<-ctx.Done()
		ln.Close()
	}()
	for {
		conn, err := ln.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				s.shutdownLogout(context.Background())
				return nil
			default:
			}
			return fmt.Errorf("accept: %w", err)
		}
		go s.handleConn(ctx, conn)
	}
}

func (s *Server) handleConn(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	line, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil && strings.TrimSpace(line) == "" {
		fmt.Fprintln(conn, "err read:", err)
		return
	}
	reply := s.Handle(ctx, strings.TrimSpace(line))
	fmt.Fprintln(conn, reply)
}

func (s *Server) Handle(ctx context.Context, cmd string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	switch {
	case cmd == "disconnect":
		return s.onDisconnect(ctx)
	case cmd == "stop":
		if s.cancel != nil {
			s.cancel()
		}
		return "ok stopping"
	case strings.HasPrefix(cmd, "connect "):
		return s.onConnect(ctx, strings.TrimPrefix(cmd, "connect "))
	default:
		return "err unknown command"
	}
}

func (s *Server) onConnect(ctx context.Context, ssid string) string {
	cfg, st, err := s.load()
	if err != nil {
		return "err load: " + err.Error()
	}
	if cfg.Paused {
		return "skip paused"
	}
	if !hasSSID(cfg.SSIDs, ssid) {
		return "skip unregistered"
	}
	select {
	case <-ctx.Done():
		return "skip shutting down"
	case <-time.After(s.SettleDelay):
	}

	online, err := s.Portal.Livecheck(ctx)
	if err == nil && online {
		s.Notify(fmt.Sprintf("Already online on %s — no action needed", ssid))
		s.logf("already online on %s", ssid)
		return "ok already-online"
	}
	if st.IsRecent(state.ActionLogin, cfg.Timings.LoginCooldown) {
		return "skip login-cooldown"
	}
	username, password, err := s.creds(cfg)
	if err != nil {
		s.Notify("Sign-in failed: no credentials stored — set them up in the app")
		s.logf("creds unavailable: %v", err)
		return "err no-creds"
	}
	ok, msg, err := s.Portal.Login(ctx, username, password)
	switch {
	case err != nil:
		s.Notify(fmt.Sprintf("Sign-in failed on %s: %v", ssid, err))
		s.logf("login error: %v", err)
		return "err login"
	case ok:
		bssid, _ := s.Wifi.ActiveBSSID()
		saveState(s.StatePath, state.ActionLogin, bssid)
		s.Notify(fmt.Sprintf("Signed in on %s", ssid))
		s.logf("login OK on %s (bssid %s)", ssid, bssid)
		return "ok logged-in"
	case isLimitMessage(msg):
		s.Notify(fmt.Sprintf(
			"Login limit reached on %s — your session is still bound to the AP you last connected through "+
				"(the portal can see all sessions but can't move one between APs — known portal bug). "+
				"Move back into that AP's range, or wait out the session timeout (~30 min), and try again.", ssid))
		s.logf("login limit reached on %s: %s", ssid, msg)
		return "ok login-limit"
	default:
		s.Notify(fmt.Sprintf("Sign-in failed on %s: %s", ssid, msg))
		s.logf("login rejected on %s: %s", ssid, msg)
		return "err rejected"
	}
}

func (s *Server) onDisconnect(ctx context.Context) string {
	cfg, st, err := s.load()
	if err != nil {
		return "err load: " + err.Error()
	}
	if st.IsRecent(state.ActionLogout, cfg.Timings.LogoutCooldown) {
		return "skip logout-cooldown"
	}
	username, _, err := s.creds(cfg)
	if err != nil {
		return "err no-creds"
	}
	if err := s.Portal.Logout(ctx, username); err != nil {
		s.logf("logout error (best-effort): %v", err)
		return "err logout"
	}
	saveState(s.StatePath, state.ActionLogout, "")
	s.logf("logout OK after disconnect")
	return "ok logged-out"
}

func (s *Server) shutdownLogout(ctx context.Context) {
	cfg, st, err := s.load()
	if err != nil || st.IsRecent(state.ActionLogout, cfg.Timings.LogoutCooldown) {
		return
	}
	up, err := s.Wifi.Up()
	if err != nil || !up {
		s.logf("shutdown without network — skipping logout")
		return
	}
	username, _, err := s.creds(cfg)
	if err != nil {
		return
	}
	cctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := s.Portal.Logout(cctx, username); err != nil {
		s.logf("shutdown logout failed (best-effort): %v", err)
		return
	}
	saveState(s.StatePath, state.ActionLogout, "")
	s.logf("clean shutdown logout OK")
}

func (s *Server) load() (*config.Config, *state.State, error) {
	cfg, err := config.Load(s.ConfigPath)
	if err != nil && err != config.ErrNoConfig {
		return nil, nil, err
	}
	st, err := state.Load(s.StatePath)
	if err != nil && !errors.Is(err, state.ErrNoState) {
		return nil, nil, err
	}
	return cfg, st, nil
}

func (s *Server) creds(cfg *config.Config) (string, string, error) {
	fp, err := config.MachineFingerprint()
	if err != nil {
		return "", "", err
	}
	return cfg.GetCreds(fp)
}

func hasSSID(list []string, ssid string) bool {
	for _, s := range list {
		if s == ssid {
			return true
		}
	}
	return false
}

func isLimitMessage(msg string) bool {
	return strings.Contains(strings.ToLower(msg), "maximum login limit") ||
		strings.Contains(strings.ToLower(msg), "login limit")
}

func saveState(path, action, bssid string) {
	st := &state.State{Action: action, Timestamp: time.Now().Unix(), BSSID: bssid}
	if err := state.Save(path, st); err != nil {
		log.Printf("state save: %v", err)
	}
}

func (s *Server) logf(format string, args ...any) {
	f, err := os.OpenFile(filepath.Join(filepath.Dir(s.ConfigPath), "log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return
	}
	defer f.Close()
	fmt.Fprintf(f, "%s %s\n", time.Now().Format("2006-01-02 15:04:05"), fmt.Sprintf(format, args...))
}
