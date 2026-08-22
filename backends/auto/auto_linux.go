//go:build linux

package auto

import (
	"github.com/achar-pranav/captive-bypass/backends"
	"github.com/achar-pranav/captive-bypass/backends/nmcli"
)

func Default() backends.Backend {
	return nmcli.New()
}
