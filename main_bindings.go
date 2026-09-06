//go:build bindings || desktop

package main

import "github.com/achar-pranav/captive-bypass/internal/gui"

// defaultEntrypoint is called by main when the process is started with no
// command line arguments. Under `bindings` or `desktop` (wails build) build tags,
// it launches the GUI directly (enabling native .app / .exe launch without CLI args).
// The CLI entry point in main.go still handles every explicit command.
func defaultEntrypoint() {
	if err := gui.Run(); err != nil {
		panic(err)
	}
}
