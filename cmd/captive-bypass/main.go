package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/achar-pranav/captive-bypass/internal/config"
	"github.com/achar-pranav/captive-bypass/internal/gui"
	"github.com/achar-pranav/captive-bypass/internal/install"
	"github.com/achar-pranav/captive-bypass/internal/portal"
	"github.com/achar-pranav/captive-bypass/internal/serve"
	"github.com/achar-pranav/captive-bypass/internal/state"
)

const usageText = `captive-bypass - PESU Sophos/Cyberoam captive portal auto-login
Usage:
  captive-bypass <command> [flags]

Commands (mirrors of the original script):
  login            force a login attempt now (best-effort logout first)
  logout           send a logout request to clear the session
  enable           resume auto login/logout
  disable          pause auto login/logout (e.g. a friend's login)
  update-creds     store a credential set: --user SRN --pass PASS [--name NAME]
  set-network      replace recognized networks: set-network SSID [SSID...]
  install          set up the background watcher (no admin needed)
  uninstall        remove the watcher AND wipe credentials, config, state

Watcher/service:
  serve            background watcher (kernel events on linux)
  watch            (windows) WLAN event listener -> running watcher
  event connect <ssid> | event disconnect    poke a running watcher manually

  gui              control panel
  dev              tester helpers: wipe | reset-state | clear-vanguard | force
  -h, --help       show this help

Environment overrides (advanced):
  CAPTIVE_BYPASS_PORTAL      portal base URL (default https://rr.pes.edu:8090)
  CAPTIVE_BYPASS_CONFIG      config dir (default ~/.config/captive-bypass)
  CAPTIVE_BYPASS_RETRY_DELAY seconds between retries (default 5)
  CAPTIVE_DEBOUNCE_MS        netlink debounce ms (linux watcher, default 800)
`

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}
	arg := strings.TrimPrefix(os.Args[1], "--")
	switch arg {
	case "-h", "help":
		usage()
	case "login":
		runLogin()
	case "logout":
		runLogout()
	case "enable", "disable":
		runToggle(arg == "disable")
	case "update-creds":
		runUpdateCreds(os.Args[2:])
	case "set-network":
		runSetNetwork(os.Args[2:])
	case "serve":
		runServe()
	case "event":
		runEvent(os.Args[2:])
	case "watch":
		runWatch()
	case "install":
		must(install.Enable(), "install")
		fmt.Println("Watcher installed.")
	case "uninstall":
		if !confirm("Remove watcher AND wipe credentials, config, state? [y/N] ") {
			return
		}
		must(install.Uninstall(), "uninstall")
		fmt.Println("Uninstalled: watcher removed, config wiped.")
	case "dev":
		runDev(os.Args[2:])
	case "gui":
		if err := gui.Run(); err != nil {
			fmt.Fprintln(os.Stderr, "captive-bypass:", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "captive-bypass: unknown command %q\n", os.Args[1])
		os.Exit(2)
	}
}

func loadConfigOrDie() *config.Config {
	dir := config.DefaultDir()
	cfg, err := config.Load(filepath.Join(dir, "config.json"))
	if err != nil && err != config.ErrNoConfig {
		must(err, "load config")
	}
	return cfg
}

func saveConfig(cfg *config.Config) {
	must(config.Save(filepath.Join(config.DefaultDir(), "config.json"), cfg), "save config")
}

func runLogin() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cfg := loadConfigOrDie()
	fp := fingerprint()
	user, pass, err := cfg.GetActiveCreds(fp)
	if err != nil {
		must(err, "credentials")
	}
	p := portal.New(cfg.Portal, nil)
	_ = p.Logout(ctx, user)
	time.Sleep(500 * time.Millisecond)
	ok, msg, err := p.Login(ctx, user, pass)
	switch {
	case err != nil:
		must(err, "login")
	case ok:
		fmt.Println("Logged in.")
	default:
		fmt.Fprintf(os.Stderr, "Portal says: %s\n", msg)
		os.Exit(1)
	}
}

func runLogout() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cfg := loadConfigOrDie()
	user, _, err := cfg.GetActiveCreds(fingerprint())
	if err != nil {
		user = ""
	}
	must(portal.New(cfg.Portal, nil).Logout(ctx, user), "logout")
	fmt.Println("Logged out.")
}

func runToggle(disable bool) {
	cfg := loadConfigOrDie()
	cfg.Paused = disable
	saveConfig(cfg)
	if disable {
		fmt.Println("Auto-login paused (master switch off).")
		return
	}
	fmt.Println("Auto-login enabled.")
}

func runUpdateCreds(args []string) {
	var user, pass, name string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--user":
			i++
			if i < len(args) {
				user = args[i]
			}
		case "--pass":
			i++
			if i < len(args) {
				pass = args[i]
			}
		case "--name":
			i++
			if i < len(args) {
				name = args[i]
			}
		}
	}
	if user == "" || pass == "" {
		fmt.Fprintln(os.Stderr, "usage: captive-bypass update-creds --user SRN --pass PASSWORD [--name NAME]")
		os.Exit(2)
	}
	cfg := loadConfigOrDie()
	must(cfg.SetCredSet(fingerprint(), name, user, pass), "store credentials")
	saveConfig(cfg)
	fmt.Printf("Credential set %q stored.\n", name)
}

func runSetNetwork(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: captive-bypass set-network SSID [SSID...]")
		os.Exit(2)
	}
	cfg := loadConfigOrDie()
	cfg.SSIDs = args
	saveConfig(cfg)
	fmt.Printf("Registered networks: %s\n", strings.Join(args, ", "))
}

func fingerprint() []byte {
	fp, err := config.MachineFingerprint()
	must(err, "machine fingerprint")
	return fp
}

func must(err error, what string) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "captive-bypass:", what+":", err)
		os.Exit(1)
	}
}

func confirm(prompt string) bool {
	fmt.Print(prompt)
	var ans string
	if _, err := fmt.Scanln(&ans); err != nil {
		return false
	}
	switch strings.ToLower(ans) {
	case "y", "yes":
		return true
	}
	return false
}

func runDev(args []string) {
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "usage: captive-bypass dev wipe|reset-state|clear-vanguard|force")
		os.Exit(2)
	}
	switch args[0] {
	case "wipe":
		must(os.RemoveAll(config.DefaultDir()), "wipe")
		fmt.Println("Wiped config, credentials, state.")
	case "reset-state":
		err := os.Remove(state.DefaultPath())
		if err != nil && !os.IsNotExist(err) {
			must(err, "reset-state")
		}
		fmt.Println("State cleared.")
	case "clear-vanguard":
		p := filepath.Join(config.DefaultDir(), "vanguard.log")
		err := os.Remove(p)
		if err != nil && !os.IsNotExist(err) {
			must(err, "clear-vanguard")
		}
		fmt.Println("Vanguard log cleared (if any).")
	case "force":
		st := &state.State{}
		must(state.Save(state.DefaultPath(), st), "force")
		fmt.Println("Cooldowns bypassed for the next event.")
	default:
		fmt.Fprintf(os.Stderr, "captive-bypass: unknown dev command %q\n", args[0])
		os.Exit(2)
	}
}

func runServe() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := serve.New().Run(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "captive-bypass:", err)
		os.Exit(1)
	}
}



func runEvent(args []string) {
	var cmd string
	switch {
	case len(args) == 2 && args[0] == "connect":
		cmd = "connect " + args[1]
	case len(args) == 1 && args[0] == "disconnect":
		cmd = "disconnect"
	default:
		fmt.Fprintln(os.Stderr, `usage: captive-bypass event connect <ssid> | event disconnect`)
		os.Exit(2)
	}
	conn, err := net.Dial("unix", serve.DefaultSocketPath())
	if err != nil {
		fmt.Fprintln(os.Stderr, "captive-bypass: watcher not running:", err)
		os.Exit(1)
	}
	defer conn.Close()
	fmt.Fprintln(conn, cmd)
	io.Copy(os.Stdout, conn)
}

func usage() {
	fmt.Print(usageText)
}
