package backends

import "sort"

type AP struct {
	SSID    string
	BSSID   string
	Signal  int
	Secured bool
}

type Backend interface {
	ActiveSSID() (string, error)
	ActiveBSSID() (string, error)
	Signal() (int, error)
	Up() (bool, error)
	Scan() ([]AP, error)
}

func Consolidate(aps []AP) []AP {
	best := make(map[string]AP, len(aps))
	for _, ap := range aps {
		if ap.SSID == "" {
			continue
		}
		if cur, ok := best[ap.SSID]; !ok || ap.Signal > cur.Signal {
			best[ap.SSID] = ap
		}
	}
	out := make([]AP, 0, len(best))
	for _, ap := range best {
		out = append(out, ap)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Signal != out[j].Signal {
			return out[i].Signal > out[j].Signal
		}
		return out[i].SSID < out[j].SSID
	})
	return out
}
