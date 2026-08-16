//go:build windows

package config

import "golang.org/x/sys/windows/registry"

func platformFingerprint() ([]byte, error) {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Cryptography`, registry.QUERY_VALUE)
	if err != nil {
		return nil, err
	}
	defer k.Close()
	v, _, err := k.GetStringValue("MachineGuid")
	if err != nil {
		return nil, err
	}
	return []byte(v), nil
}
