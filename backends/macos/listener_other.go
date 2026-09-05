//go:build !darwin

package macos

import (
	"context"
	"errors"
)

var ErrNoListener = errors.New("corewlan event listener requires macOS")

func Listen(ctx context.Context, sockPath string) error {
	return ErrNoListener
}
