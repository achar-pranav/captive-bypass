//go:build darwin

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/achar-pranav/captive-bypass/backends/macos"
	"github.com/achar-pranav/captive-bypass/internal/serve"
)

func runWatch() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := macos.Listen(ctx, serve.DefaultSocketPath()); err != nil {
		fmt.Fprintln(os.Stderr, "captive-bypass:", err)
		os.Exit(1)
	}
}
