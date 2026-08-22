package iw

import (
	"testing"
)

const linkSample = `Connected to 9c:53:22:5e:12:a0 (on wlan0)
	SSID: ELEMENT BLOCK
	freq: 5180
	signal: -51 dBm
	tx bitrate: 866.7 MBit/s VHT-MCS 8 80MHz
	bss flags: short-preamble short-slot-time
	dtim period: 2
	beacon int: 100
`

func TestParseLink(t *testing.T) {
	if got := parseLinkField(linkSample, "ssid"); got != "ELEMENT BLOCK" {
		t.Errorf("ssid = %q", got)
	}
	bssid := parseConnectedBSSID(linkSample)
	if bssid != "9c:53:22:5e:12:a0" {
		t.Errorf("bssid = %q", bssid)
	}
	sig, err := parseSignal(linkSample)
	if err != nil || sig != -51 {
		t.Errorf("signal = %d, %v; want -51, nil", sig, err)
	}
}

func TestParseSignalMissing(t *testing.T) {
	if _, err := parseSignal("Not connected.\n"); err == nil {
		t.Error("expected error for not-connected output")
	}
}
