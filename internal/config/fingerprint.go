package config

func MachineFingerprint() ([]byte, error) {
	return platformFingerprint()
}
