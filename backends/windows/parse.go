package windows

import (
	"strconv"
	"strings"
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
