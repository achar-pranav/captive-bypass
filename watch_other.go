//go:build !darwin && !windows

package main

import (
	"fmt"
	"os"
)

func runWatch() {
	fmt.Fprintln(os.Stderr, "captive-bypass: standalone watch command not needed on Linux (serve uses kernel watcher)")
	os.Exit(1)
}
