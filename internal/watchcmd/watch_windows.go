//go:build windows

package watchcmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/achar-pranav/captive-bypass/backends/windows"
	"github.com/achar-pranav/captive-bypass/internal/serve"
)

func Run() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := windows.Listen(ctx, serve.DefaultSocketPath()); err != nil {
		fmt.Fprintln(os.Stderr, "captive-bypass:", err)
		os.Exit(1)
	}
}
