//go:build !bindings && !desktop

package entrypoint

import "fmt"
import "os"

// defaultEntrypoint is called by main when the process is started with no
// command line arguments in standard CLI builds. It prints usage and exits.
func DefaultEntrypoint(usage func()) {
	usage()
	os.Exit(1)
}
