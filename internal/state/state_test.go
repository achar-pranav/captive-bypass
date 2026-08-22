package state

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state")
	want := &State{Action: ActionLogin, Timestamp: 1723740000, BSSID: "aa:bb:cc:dd:ee:ff"}
	if err := Save(path, want); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Load = %+v, want %+v", got, want)
	}
}

func TestMissingState(t *testing.T) {
	_, err := Load(filepath.Join(t.TempDir(), "nope"))
	if !errors.Is(err, ErrNoState) {
		t.Errorf("Load = %v, want ErrNoState", err)
	}
}

func TestCorruptState(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state")
	if err := os.WriteFile(path, []byte("login 1723740000"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if _, err := Load(path); err == nil {
		t.Error("Load of garbage succeeded, want error")
	}
}

func TestSavePerms(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state")
	if err := Save(path, &State{Action: ActionLogout}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	fi, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if fi.Mode().Perm() != 0o600 {
		t.Errorf("file mode = %o, want 600", fi.Mode().Perm())
	}
}

func TestIsRecent(t *testing.T) {
	restore := freezeNow(1723740000)
	defer restore()

	tests := []struct {
		name  string
		state *State
		check string
		with  int
		want  bool
	}{
		{"fresh login", &State{Action: ActionLogin, Timestamp: 1723740000 - 30}, ActionLogin, 60, true},
		{"exactly at limit", &State{Action: ActionLogin, Timestamp: 1723740000 - 60}, ActionLogin, 60, false},
		{"expired", &State{Action: ActionLogin, Timestamp: 1723740000 - 61}, ActionLogin, 60, false},
		{"wrong action", &State{Action: ActionLogout, Timestamp: 1723740000 - 5}, ActionLogin, 60, false},
		{"zero timestamp", &State{Action: ActionLogin}, ActionLogin, 60, false},
		{"nil state", nil, ActionLogin, 60, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.state.IsRecent(tt.check, tt.with); got != tt.want {
				t.Errorf("IsRecent(%q, %d) = %v, want %v", tt.check, tt.with, got, tt.want)
			}
		})
	}
}

func freezeNow(ts int64) func() {
	saved := now
	now = func() int64 { return ts }
	return func() { now = saved }
}
