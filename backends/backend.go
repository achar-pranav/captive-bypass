package backends

type Backend interface {
	ActiveSSID() (string, error)
	ActiveBSSID() (string, error)
	Signal() (int, error)
	Up() (bool, error)
}
