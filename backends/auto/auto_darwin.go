//go:build darwin

package auto

import (
	"github.com/achar-pranav/captive-bypass/backends"
	"github.com/achar-pranav/captive-bypass/backends/macos"
)

func Default() backends.Backend {
	return macos.New()
}
