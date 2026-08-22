//go:build !linux && !windows

package auto

import "github.com/achar-pranav/captive-bypass/backends"

func Default() backends.Backend {
	return nil
}
