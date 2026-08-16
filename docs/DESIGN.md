# DESIGN.md — design notes & research log

Human-facing context. Agents: read this only when an issue references it.
This is the "why" behind the code; the "what" lives in GitHub issues.

## The problem (ELI5)
Campus WiFi is many small radios (Access Points / APs) sharing one network name
(SSID), spread across buildings and floors. The login system (Sophos/Cyberoam
captive portal) remembers "this laptop is logged in over here." When you walk
from floor 1 to floor 4 — same SSID, different AP — the old "logged in here"
record is never deleted. So when the laptop tries to log in at the new AP, the
portal says: "maximum login limit reached." You're not actually limited; the
portal just thinks you're still logged in at the old spot.

## The one fact everything hangs on
A logout request (POST logout.xml, mode=193) ONLY releases the session if it is
sent from the SAME location (AP/switch) the session was created on. Sending it
from the new AP does nothing. (Verified empirically: the user had to walk back,
reconnect to the old AP, send logout, then return.)

## Session timeout
Worst case after a missed logout: the portal keeps the session until its idle
timeout. Rough estimate ~30 Minutes; NOT measured. This defines the worst-case
lockout window for a hard radio drop.

## The tiered mitigation (why the roam fix is "best effort")
1. Pre-emptive logout on signal degradation — fire while the RSSI ramp is
   happening and we still have a usable link (a walk gives seconds of warning).
2. NM dispatcher pre-down logout — catches graceful disassociates (coverage
   edge) where the client initiates the drop.
3. Hard fade (elevator/corner) — logout never ships; accept lockout until the
   ~30min session timeout. This is NOT recoverable client-side. Do not over-promise.

## Validation: empirical only
No lab. The user refuses to press "disconnect" in real life — the tool must be
validated by walking around campus (classroom -> classroom, building ->
building). Test log below.

### Test log
(date, route, RSSI samples, BSSID changes, logout fired? result)

## The nuclear option (flagged, not chosen)
Randomize the MAC per connect => every attach is a "fresh device", so a stale
binding can never bite. Costs: re-auth every connect; looks like device churn
to IT (policy risk). Kept as a documented escape hatch only.

## GUI framework choice
Chosen: **Fyne** — pure Go native window, no browser/webview, RAM footprint
~5-15 MB (budget: <=20 MB). Needs CGO (OpenGL) to draw, so each OS builds its
own copy (Linux on Linux, Windows on Windows) — the accepted tradeoff for a
small footprint.

Alternatives considered and rejected:
- **Wails** (Go backend + native webview, HTML/CSS/JS frontend): runtime
  footprint of a full browser engine (~100-300 MB RAM), CGO + per-OS webview
  deps (WebKitGTK, WebView2), complicates cross-compiling. Rejected: footprint
  too big for a "mini helper app".
- **Embedded local web UI** (Go serves HTML/CSS/JS on localhost, browser tab):
  pure Go, no CGO, cross-compiles all targets from one machine. Rejected: the
  user wants a native mini-app, not a browser tab.
- **Gio** (pure Go immediate-mode UI): small footprint, but lower-level and
  more work than a simple control panel needs. Rejected: Fyne is simpler.
- **Electron**: way too heavy and not Go. Rejected.

## ELI5 glossary
| Term | Plain meaning |
|---|---|
| Captive portal | Login page a network forces before letting you use the internet |
| SSID | The network name you see and pick (e.g. "ELEMENT BLOCK") |
| BSSID | Unique hardware ID of one specific AP radio |
| Access Point (AP) | The physical WiFi box you connect to |
| RSSI / signal | How strong the connection is (dBm; more negative = weaker) |
| MAC address | Hardware ID of your network card |
| Session | The portal's record of "this device is allowed online" |
| DHCP | Automatic process of getting an IP address on a network |
| Switch / VLAN | Network plumbing grouping APs; the "location" the portal keys on |
| nmcli | Command-line tool for NetworkManager (Linux WiFi manager) |
| NetworkManager | Daemon managing WiFi/ethernet on most Linux desktops |
| iwd / wpa_supplicant | Alternative Linux WiFi managers (not supported yet) |
| Dispatcher | NM hook scripts that run on WiFi events (up/down) |
| curl | Command-line tool that sends HTTP requests |
| Epoch milliseconds | Milliseconds since 1970-01-01; Sophos wants it as a timestamp field |
| Polling | Repeatedly checking state (vs. waiting for an event) |
