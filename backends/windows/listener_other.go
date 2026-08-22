//go:build !windows || (!amd64 && !arm64)

package windows

import (
	"context"
	"errors"
)

var ErrNoListener = errors.New("WLAN event listener requires windows amd64/arm64")

func Listen(ctx context.Context, sockPath string) error {
	return ErrNoListener
}
