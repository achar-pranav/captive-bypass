package nmcli_test

import (
	"github.com/achar-pranav/captive-bypass/backends"
	"github.com/achar-pranav/captive-bypass/backends/nmcli"
)

var _ backends.Backend = (*nmcli.Backend)(nil)
