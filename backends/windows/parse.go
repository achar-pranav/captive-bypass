package windows

import (
	"strconv"
	"strings"

	"github.com/achar-pranav/captive-bypass/backends"
)

type snapshot struct {
	SSID   string
	BSSID  string
	Signal int
	Up     bool
}

func parseInterfaces(out string) snapshot {
	var snap snapshot
	for _, line := range strings.Split(out, "\n") {
		i := strings.Index(line, ":")
		if i < 0 {
			continue
		}
		key := strings.TrimSpace(line[:i])
		val := strings.TrimSpace(strings.TrimSuffix(line[i+1:], "\r"))
		switch strings.ToUpper(key) {
		case "SSID":
			snap.SSID = val
			snap.Up = true
		case "BSSID":
			snap.BSSID = val
			snap.Up = true
		default:
			if strings.HasSuffix(val, "%") {
				snap.Signal = parsePercent(val)
			}
		}
	}
	return snap
}

func parsePercent(val string) int {
	n, err := strconv.Atoi(strings.TrimSuffix(strings.TrimSpace(val), "%"))
	if err != nil {
		return -1
	}
	return n
}

type netScan struct {
	SSID    string
	auth    string
	BSSID   string
	pending bool
}

func parseNetworks(out string) []backends.AP {
	var aps []backends.AP
	var cur netScan
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSuffix(line, "\r")
		i := strings.Index(line, ":")
		if i < 0 {
			continue
		}
		key := strings.TrimSpace(line[:i])
		val := strings.TrimSpace(line[i+1:])
		up := strings.ToUpper(key)
		switch {
		case up == "SSID" || strings.HasPrefix(up, "SSID "):
			cur = netScan{SSID: val}
		case up == "AUTHENTICATION":
			cur.auth = val
		case up == "BSSID" || strings.HasPrefix(up, "BSSID "):
			cur.BSSID = val
			cur.pending = true
		default:
			if !cur.pending || !strings.HasSuffix(val, "%") {
				continue
			}
			aps = append(aps, backends.AP{
				SSID:    cur.SSID,
				BSSID:   cur.BSSID,
				Signal:  parsePercent(val),
				Secured: cur.auth != "" && !strings.EqualFold(cur.auth, "open"),
			})
			cur.pending = false
		}
	}
	return aps
}
