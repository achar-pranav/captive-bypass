package nmcli

import (
	"reflect"
	"testing"

	"github.com/achar-pranav/captive-bypass/backends"
)

func TestParseScan(t *testing.T) {
	out := "ELEMENT BLOCK\\ 2:AA\\:BB\\:CC\\:DD\\:EE\\:FF:92:wpa2\n" +
		"PESU-GUEST:11\\:22\\:33\\:44\\:55\\:66:40:--\n" +
		"PESU-IOT:11\\:22\\:33\\:44\\:55\\:77:55:\n" +
		"\n"
	got := parseScan(out)
	want := []backends.AP{
		{SSID: "ELEMENT BLOCK 2", BSSID: "AA:BB:CC:DD:EE:FF", Signal: 92, Secured: true},
		{SSID: "PESU-GUEST", BSSID: "11:22:33:44:55:66", Signal: 40},
		{SSID: "PESU-IOT", BSSID: "11:22:33:44:55:77", Signal: 55},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parseScan:\n got %+v\nwant %+v", got, want)
	}
}

func TestConsolidate(t *testing.T) {
	in := []backends.AP{
		{SSID: "weak", Signal: 10},
		{SSID: "campus", BSSID: "1", Signal: 50},
		{SSID: "", Signal: 99},
		{SSID: "campus", BSSID: "2", Signal: 80},
	}
	got := backends.Consolidate(in)
	want := []backends.AP{
		{SSID: "campus", BSSID: "2", Signal: 80},
		{SSID: "weak", Signal: 10},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Consolidate:\n got %+v\nwant %+v", got, want)
	}
}
