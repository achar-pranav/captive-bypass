//go:build linux

package backends

import "github.com/achar-pranav/captive-bypass/backends/nmcli"

func Default() Backend {
	return nmcli.New()
}
