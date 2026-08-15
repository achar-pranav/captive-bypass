package main

import (
	"fmt"
	"os"
)

const usageText = `captive-bypass - PESU Sophos/Cyberoam captive portal auto-login

Usage:
  captive-bypass <command>

Commands:
  login     log in to the captive portal
  logout    log out of the captive portal
  serve     headless poller (auto-login on trigger SSID)
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
	case "login", "logout", "serve", "gui":
		notImplemented(os.Args[1])
	default:
		fmt.Fprintf(os.Stderr, "captive-bypass: unknown command %q\n", os.Args[1])
		os.Exit(2)
	}
}

func usage() {
	fmt.Print(usageText)
}

func notImplemented(cmd string) {
	fmt.Fprintf(os.Stderr, "captive-bypass: %s: not yet implemented\n", cmd)
	os.Exit(1)
}
