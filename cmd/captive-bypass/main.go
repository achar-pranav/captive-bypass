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

	"github.com/achar-pranav/captive-bypass/backends/windows"
	"github.com/achar-pranav/captive-bypass/internal/config"
	"github.com/achar-pranav/captive-bypass/internal/gui"
	"github.com/achar-pranav/captive-bypass/internal/install"
	"github.com/achar-pranav/captive-bypass/internal/serve"
	"github.com/achar-pranav/captive-bypass/internal/state"
)

const usageText = `captive-bypass - PESU Sophos/Cyberoam captive portal auto-login
Usage:
  captive-bypass <command>

Commands:
  login     log in to the captive portal
  logout    log out of the captive portal
  serve     background watcher: auto sign-in/out on WiFi events (no polling)
  watch     (windows) WLAN event listener; forwards events to a running watcher
  event     send an event to a running watcher:
            'event connect <ssid>' | 'event disconnect'
  gui       control panel
  install   set up the background watcher (service + hook/task)
  uninstall remove the watcher AND wipe credentials, config, state
  dev       tester helpers: wipe | reset-state | clear-vanguard | force

Flags:
  -h, --help  show this help
`

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}
	switch os.Args[1] {
	case "-h", "--help":
		usage()
	case "login", "logout":
		notImplemented(os.Args[1])
	case "serve":
		runServe()
	case "event":
		runEvent(os.Args[2:])
	case "watch":
		runWatch()
	case "install":
		must(install.Enable(), "install")
		fmt.Println("Watcher installed: service + dispatcher hook active.")
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

func runWatch() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := windows.Listen(ctx, serve.DefaultSocketPath()); err != nil {
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

func notImplemented(cmd string) {
	fmt.Fprintf(os.Stderr, "captive-bypass: %s: not yet implemented\n", cmd)
	os.Exit(1)
}
