package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/achar-pranav/captive-bypass/internal/gui"
	"github.com/achar-pranav/captive-bypass/internal/serve"
)

const usageText = `captive-bypass - PESU Sophos/Cyberoam captive portal auto-login

Usage:
  captive-bypass <command>

Commands:
  login     log in to the captive portal
  logout    log out of the captive portal
  serve     background watcher: auto sign-in/out on WiFi events (no polling)
  event     send an event to a running watcher:
            'event connect <ssid>' | 'event disconnect'
  gui       control panel

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

func notImplemented(cmd string) {
	fmt.Fprintf(os.Stderr, "captive-bypass: %s: not yet implemented\n", cmd)
	os.Exit(1)
}
