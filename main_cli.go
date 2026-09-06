//go:build !bindings && !desktop

package main

import "os"

// defaultEntrypoint is called by main when the process is started with no
// command line arguments in standard CLI builds. It prints usage and exits.
func defaultEntrypoint() {
	usage()
	os.Exit(1)
}
