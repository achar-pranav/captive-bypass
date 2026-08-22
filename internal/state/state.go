package state

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/achar-pranav/captive-bypass/internal/config"
)

const (
	ActionLogin  = "login"
	ActionLogout = "logout"
)

var ErrNoState = errors.New("no state file")

var now = func() int64 { return time.Now().Unix() }

type State struct {
	Action    string `json:"action"`
	Timestamp int64  `json:"timestamp"`
	BSSID     string `json:"bssid"`
}

func DefaultPath() string {
	return filepath.Join(config.DefaultDir(), "state")
}

func Load(path string) (*State, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNoState
		}
		return nil, err
	}
	var s State
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func Save(path string, s *State) error {
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	if err := os.Chmod(tmp, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func (s *State) IsRecent(action string, within int) bool {
	if s == nil || s.Action != action || s.Timestamp <= 0 {
		return false
	}
	return now()-s.Timestamp < int64(within)
}
