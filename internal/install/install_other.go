//go:build !linux && !windows

package install

import "errors"

var ErrUnsupported = errors.New("install not supported on this OS yet")

func Enable() error         { return ErrUnsupported }
func Disable() error        { return ErrUnsupported }
func Status() (bool, error) { return false, ErrUnsupported }
func Uninstall() error      { return ErrUnsupported }
