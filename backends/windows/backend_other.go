//go:build !windows

package windows

import (
	"errors"

	"github.com/achar-pranav/captive-bypass/backends"
)

var ErrUnsupported = errors.New("netsh backend requires windows")

type Backend struct{}

func New() *Backend { return &Backend{} }

func (b *Backend) ActiveSSID() (string, error)  { return "", ErrUnsupported }
func (b *Backend) ActiveBSSID() (string, error) { return "", ErrUnsupported }
func (b *Backend) Signal() (int, error)         { return 0, ErrUnsupported }
func (b *Backend) Up() (bool, error)            { return false, ErrUnsupported }
func (b *Backend) Scan() ([]backends.AP, error) { return nil, ErrUnsupported }
