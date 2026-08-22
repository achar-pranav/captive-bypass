//go:build windows

package auto

import (
	"github.com/achar-pranav/captive-bypass/backends"
	winwifi "github.com/achar-pranav/captive-bypass/backends/windows"
)

func Default() backends.Backend {
	return winwifi.New()
}
