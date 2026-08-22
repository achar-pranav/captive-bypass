//go:build windows

package backends

import winwifi "github.com/achar-pranav/captive-bypass/backends/windows"

func Default() Backend {
	return winwifi.New()
}
