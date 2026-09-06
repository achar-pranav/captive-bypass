//go:build !bindings

package main

import "os"

// defaultEntrypoint is called by main when the process is started with no
// command line arguments. In normal builds this prints usage and exits (the
// CLI requires an explicit verb); under the `bindings` build tag it instead
// launches the GUI so Wails can generate frontend bindings (see
// main_bindings.go).
func defaultEntrypoint() {
	usage()
	os.Exit(1)
}
