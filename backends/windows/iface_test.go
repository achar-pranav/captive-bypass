package windows_test

import (
	"testing"

	"github.com/achar-pranav/captive-bypass/backends"
	winwifi "github.com/achar-pranav/captive-bypass/backends/windows"
)

var _ backends.Backend = (*winwifi.Backend)(nil)

func TestStubOnNonWindows(t *testing.T) {
	b := winwifi.New()
	if _, err := b.ActiveSSID(); err == nil {
		t.Error("ActiveSSID should error on non-windows builds")
	}
}
