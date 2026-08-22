package nmcli

import (
	"errors"
	"testing"
)

func TestParseActiveSSID(t *testing.T) {
	out := "no:PESU-GUEST\nno:FreeWiFi\nyes:ELEMENT BLOCK\nno:Neighbours 2.4G\n"
	if got := parseActive(out); got != "ELEMENT BLOCK" {
		t.Errorf("parseActive = %q, want %q", got, "ELEMENT BLOCK")
	}
}

func TestParseActiveSSIDWithColons(t *testing.T) {
	out := "no:other\nyes:PESU-WIFI\\:STAFF\\:5G\n"
	if got := parseActive(out); got != "PESU-WIFI:STAFF:5G" {
		t.Errorf("parseActive = %q, want %q", got, "PESU-WIFI:STAFF:5G")
	}
}

func TestUnescape(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"plain", "plain"},
		{`a\:b`, "a:b"},
		{`a\\b`, `a\b`},
		{`CE\:82\:A9\:D5\:35\:8B`, "CE:82:A9:D5:35:8B"},
		{`trailing\`, `trailing\`},
	}
	for _, tt := range tests {
		if got := unescape(tt.in); got != tt.want {
			t.Errorf("unescape(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestParseActiveNone(t *testing.T) {
	for _, out := range []string{"", "\n", "no:A\nno:B\n", "--\n"} {
		if got := parseActive(out); got != "" {
			t.Errorf("parseActive(%q) = %q, want empty", out, got)
		}
	}
}

func TestParseActiveBSSID(t *testing.T) {
	out := "no:11\\:22:33\nyes:CE\\:82\\:A9\\:D5\\:35\\:8B\n"
	if got := parseActive(out); got != "CE:82:A9:D5:35:8B" {
		t.Errorf("parseActive = %q, want %q", got, "CE:82:A9:D5:35:8B")
	}
}

const iwLink = `Connected to aa:bb:cc:dd:ee:ff (on wlan0)
	SSID: ELEMENT BLOCK
	freq: 2437
	RX: 12345 bytes (678 packets)
	TX: 23456 bytes (789 packets)
	signal: -57 dBm
	tx bitrate: 72.2 MBit/s MCS 7 short GI
	bss flags:	short-slot-time
	dtim period:	2
	beacon interval:100
`

func TestParseSignal(t *testing.T) {
	got, err := parseSignal(iwLink)
	if err != nil {
		t.Fatalf("parseSignal: %v", err)
	}
	if got != -57 {
		t.Errorf("parseSignal = %d, want -57", got)
	}
}

func TestParseSignalMissing(t *testing.T) {
	out := "Not connected.\n"
	if _, err := parseSignal(out); !errors.Is(err, ErrNoSignal) {
		t.Errorf("parseSignal = %v, want ErrNoSignal", err)
	}
}

func TestParseIface(t *testing.T) {
	out := "PESU: Staff:wlan0:802-11-wireless\neth0:eth0:802-3-ethernet\n"
	if got := parseIface(out); got != "wlan0" {
		t.Errorf("parseIface = %q, want wlan0", got)
	}
}

func TestParseIfaceNone(t *testing.T) {
	for _, out := range []string{"", "eth0:eth0:802-3-ethernet\n"} {
		if got := parseIface(out); got != "" {
			t.Errorf("parseIface(%q) = %q, want empty", out, got)
		}
	}
}

func TestParseConnectivity(t *testing.T) {
	tests := []struct {
		out  string
		want bool
	}{
		{"full\n", true},
		{"limited\n", true},
		{"portal\n", true},
		{"none\n", false},
		{"unknown\n", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := parseConnectivity(tt.out); got != tt.want {
			t.Errorf("parseConnectivity(%q) = %v, want %v", tt.out, got, tt.want)
		}
	}
}
