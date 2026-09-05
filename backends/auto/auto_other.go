//go:build !linux && !windows && !darwin

package auto

import "github.com/achar-pranav/captive-bypass/backends"

func Default() backends.Backend {
	return nil
}
