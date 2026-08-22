package nmcli

import (
	"strconv"
	"strings"

	"github.com/achar-pranav/captive-bypass/backends"
)

func (b *Backend) Scan() ([]backends.AP, error) {
	out, err := run("nmcli", "-t", "-f", "ssid,bssid,signal,security", "device", "wifi", "list")
	if err != nil {
		return nil, err
	}
	return parseScan(out), nil
}

func parseScan(out string) []backends.AP {
	var aps []backends.AP
	for _, line := range splitLines(out) {
		f := splitTerse(line)
		if len(f) < 4 {
			continue
		}
		sig, err := strconv.Atoi(f[2])
		if err != nil {
			continue
		}
		aps = append(aps, backends.AP{
			SSID:    f[0],
			BSSID:   f[1],
			Signal:  sig,
			Secured: f[3] != "" && f[3] != "--",
		})
	}
	return aps
}

func splitTerse(line string) []string {
	var fields []string
	var b []byte
	flush := func() {
		fields = append(fields, string(b))
		b = b[:0]
	}
	for i := 0; i < len(line); i++ {
		c := line[i]
		if c == '\\' && i+1 < len(line) {
			i++
			b = append(b, line[i])
			continue
		}
		if c == ':' {
			flush()
			continue
		}
		b = append(b, c)
	}
	return append(fields, string(b))
}

func splitLines(out string) []string {
	out = strings.TrimSuffix(out, "\n")
	if out == "" {
		return nil
	}
	lines := strings.Split(out, "\n")
	for i, l := range lines {
		lines[i] = strings.TrimSuffix(l, "\r")
	}
	return lines
}

func ParseScan(out string) []backends.AP { return parseScan(out) }
