//go:build bindings

package main

import "github.com/achar-pranav/captive-bypass/internal/gui"

// defaultEntrypoint is called by runNoArgs when the process is started with
// no command line arguments. Wails builds the project with the `bindings` tag
// to generate frontend bindings: it compiles the project and runs the binary
// with no arguments, expecting it to reach wails.Run so the bound methods can
// be introspected. Under this tag we launch the GUI directly; the CLI entry
// point in main.go still handles every explicit command.
func defaultEntrypoint() {
	if err := gui.Run(); err != nil {
		panic(err)
	}
}
