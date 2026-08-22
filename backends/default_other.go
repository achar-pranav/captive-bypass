//go:build !linux && !windows

package backends

func Default() Backend {
	return nil
}
