//go:build !linux

package watcher

import "errors"

var ErrUnsupported = errors.New("kernel event watcher requires linux")

type Watcher struct{}

func Start(func()) (*Watcher, error) { return nil, ErrUnsupported }

func (w *Watcher) Stop() {}
