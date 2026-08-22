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

**Confirmed empirically:** BSSID changes per classroom on the campus network
(same SSID, different AP). Walking from one class to the next triggers a stale
session lockout even without changing networks. This is not an edge case — it is
the daily experience.

## Session timeout
Worst case after a missed logout: the portal keeps the session until its idle
timeout. Rough estimate ~30 Minutes; NOT measured. This defines the worst-case
lockout window for a hard radio drop.

## The tiered mitigation (why the roam fix is "best effort")
1. **Vanguard (active):** Monitor RSSI trend while associated. On strong→weak
   signal ramp, fire predictive logout while still on the old AP. If signal
   stabilizes at a weak threshold, auto re-login + GUI notification warning
   the user they are at the edge of coverage. This is the only fix for
   classroom-to-classroom roams.
2. NM dispatcher pre-down logout — catches graceful disassociates (coverage
   edge) where the client initiates the drop.
3. Hard fade (elevator/corner) — logout never ships; accept lockout until the
   ~30min session timeout. This is NOT recoverable client-side. Do not over-promise.

## Vanguard (roam handling)
**Status: confirmed necessary.** BSSID-per-classroom makes this a daily problem,
not an edge case. Active development target.

Tier 1 is the only fix for classroom-to-classroom roams, and it's Vanguard's
job: monitor signal trend + BSSID drift and fire logout BEFORE the drop.
**Known risk: rubberbanding** — false positives firing logout too eagerly.
Decision: experimental only, off by default, never shipped as stable.

**Target flow:**
1. RSSI strong → weak trend detected: fire predictive logout while still on old
   AP (signal still usable, link still alive).
2. RSSI stabilizes at weak threshold (e.g. -75 dBm): auto re-login + GUI
   notification ("edge of coverage").
3. RSSI crosses critical threshold before logout fires: hard drop, accept
   lockout until ~30min session timeout. Not recoverable client-side.

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

## Windows wifi backend decision
- **State:** `netsh wlan show interfaces` parsing, mirroring nmcli. Native
  `wlanapi` state calls rejected: more code (buffers/GUIDs/pointers) for no
  v1 gain.
- **Signal:** log netsh percentage for now. Plan: standardize to **dBm across
  all OSes** to match tuning from test data. Windows dBm extraction unsolved —
  may need an empirical conversion scale after real testing.
- **Events:** `WlanRegisterNotification` (pure Go syscall, no CGO). No poller
  and no poller fallback — revisit only if real-hardware testing proves the
  listener broken. Goal: minimal LOC.
- **Events -> serve bridge:** the listener dials the same Unix socket the Linux
  dispatcher hook uses (`connect <ssid>` / `disconnect` wire format); AF_UNIX
  ships with Windows 10+, so `serve` has zero platform branching.
- **netsh locales:** parse acronym keys (SSID/BSSID are not translated) and
  detect Signal structurally by a trailing `%` value; localized State lines are
  ignored entirely — Up() means an SSID/BSSID line was present. Fixtures cover
  en/de/fr/es plus colons inside SSIDs.

## Event plumbing v2: kernel netlink watcher (dispatcher retired)
- **State (2026-08-23):** `internal/watcher` subscribes to rtnetlink
  (LINK + IPv4 ifaddr/route groups) as an unprivileged user. Any burst ->
  800 ms debounce (`CAPTIVE_DEBOUNCE_MS` to tune on campus) -> one state read
  via `backends/iw` (`iw dev <if> link` + `/sys/class/net`) -> classified
  connect vs disconnect -> the same serve handlers the old hook fed. The Unix
  socket stays for GUI control, Windows listener, and manual pokes. Verified
  live: down -> best-effort logout; assoc+DHCP -> settle -> livecheck -> login.
- **Why:** works under ANY manager (NM/iwd/wpa_supplicant/none) because it
  listens below them all; zero privileges ever (installer needs no pkexec at
  all now); one event source instead of per-manager hooks. Old hook retired to
  `attic/dispatcher-hook.sh`, not deleted.
- **Alternatives rejected:** keeping NM dispatcher + adding iwd D-Bus +
  wpa_cli action scripts (three code paths, forever); universal poller timer —
  kernel push is free and instant, heartbeat only if field testing shows missed
  transitions; manager-specific backends per distro family — unmaintainable.
- **Known tradeoffs:** no pre-down moment (roam pre-logout stays #11 scope);
  route blips (VPNs) trigger harmless re-checks — cooldowns + silent
  already-online-within-cooldown keep it quiet; `iw` must exist (installer
  checks at Enable time with an actionable error).

## SSID picking UI (#31)
- **State:** networks are picked from a live scan as checkboxes; toggling
  edits the registered set directly. Typed entry removed entirely — campus
  SSIDs always broadcast (dev call).
- **Alternatives rejected:** type-them-manually (original wizard) — typos
  silently break roaming detection, the exact bug class this tool exists to
  kill; checkbox list + separate Apply button — two-step ritual for no safety
  gain since changes are reversible by unchecking; multi-select list widget —
  same as checkboxes but with worse touch targets.
- **Fyne gotcha (for future us):** a bare `container.NewScroll` collapses to a
  32px min inside VBox layouts and rows vanish again (#9 regression class).
  Scroll must sit in a `Border` center slot to absorb leftover height.
- **Backend note:** scan lives on the `backends.Backend` interface (`Scan()`)
  so GUI stays OS-agnostic; the shared `AP` type forced the `Default()` factory
  out of package `backends` into `backends/auto` (import cycle: drivers import
  `backends` for `AP`, so nothing in `backends` may import drivers).

## Watcher install/uninstall (v2: zero prompts)
- **Linux:** systemd **user** unit only — write file, daemon-reload, enable.
  No pkexec anywhere in the product anymore; install/uninstall never asks for
  a password (dev requirement: trust = zero auth dialogs). Installer verifies
  `iw` exists before writing anything.
- **Windows:** two per-user scheduled tasks at logon (`serve`, `watch`),
  `/RL LIMITED` so zero admin; AF_UNIX socket works unmodified on Win10+.
- **History:** the earlier design used one explicit pkexec to drop an NM
  dispatcher hook — kept working until netlink made it unnecessary.

## Credential sets + GUI direction (this pass)
- **Model:** multiple named cred sets ({name, SRN, scrypt/AES-GCM ciphertext
  keyed by machine fingerprint}); exactly ONE active set decides what any
  login uses. Legacy single-cred configs migrate into a set named `default`
  on load. SSIDs stay a flat recognized-list — they gate WHEN to log in, the
  active set decides WITH WHAT; unregistered SSIDs are inert by design.
- **Master switch:** `paused` config flag presented in GUI as "Auto Login"
  master toggle (off overrides everything). `Vanguard` bool added as a
  placeholder toggle only — no behavior wired this pass.
- **Look:** Discord-style cards on true black (#000) with electric blue
  (#00A8FF) accents via a custom fyne theme override; compact window meant to
  be opened once, configured, closed.

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
