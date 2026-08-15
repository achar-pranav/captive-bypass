# Rewrite in Go with tiered WiFi backends

**Goal:** Replace the single bash script (`captive-bypass`) with a Go binary that
keeps every current behavior byte-for-byte and adds a backend abstraction so one
binary runs on the majority of student laptops.

**Why now:** bash cannot run natively on Windows/macOS student laptops. We're
rewriting from scratch anyway, so the substrate is chosen once: Go. It
cross-compiles to a static binary with no toolchain per target
(`GOOS=windows go build`), stdlib covers HTTP/subprocess/parsing, and goroutines
make the future roam-monitor trivial. Rust was weighed and rejected: its
strengths (safety at scale, peak perf) don't apply to a program this size, and
cross-compile + learning curve would slow a solo non-expert dev down.

## Portability target (~90% of student laptops)

Coverage matrix (assumption: campus mix is Windows-heavy, some macOS, few Linux):

- Linux — nmcli (NetworkManager) backend; then iwd, then wpa_supplicant
- Windows — `netsh wlan` backend
- macOS — `airport -I` / CoreWLAN backend
- Build matrix: linux/amd64+arm64, windows/amd64+arm64, darwin/amd64+arm64.
  One static binary per target; the *same* code, only the WiFi backend differs.

## Architecture

- `main` — subcommand surface, preserves today's CLI: `login`, `logout`,
  `--set-network`, `--update-creds`, `--install`, `--enable`/`--disable`, plus
  new `serve` (background roam-monitor daemon)
- `wifi` package — backend interface: `ActiveSSID()`, `ActiveBSSID()`, `Signal()`,
  `Up()`, `Watch(events)`. Implementations: nmcli/iwd/wpa_cli/netsh/airport.
  Parsers are unit-testable with fixture output (the colon-split BSSID trap from
  the original get_active_bssid work becomes a regression test)
- `portal` package — Sophos/Cyberoam client: login.xml mode=191,
  logout.xml mode=193, username/password/a/producttype, CDATA LIVE/message
  parsing. Same curl behavior (InsecureSkipVerify kept, gated)
- `roam` package — the non-negotiable roaming fact stays: logout must fire
  BEFORE the old AP is lost, only valid from the same AP. v1 ships the
  *skeleton*: BSSID+RSSI sampling + logging, no behavior change. Hard radio
  drops remain explicitly out of scope
- `config`/`state` — keep `~/.config/captive-bypass/` layout
  (`os.UserConfigDir`), creds 0600, no secrets in logs
- Linux install path — NetworkManager dispatcher now invokes the Go binary;
  Windows/macOS get their own event hooks in follow-on issues

## Preserved behavior (regression baseline)

Retry/cooldown timings, pre-login logout + gap, disabled flag, trigger-SSID
gating, first-run setup flow, env overrides.

## Done when

- Go binary reproduces every current bash behavior (manual parity run)
- Backend interface + nmcli implementation, with unit tests on parsed fixtures
  (SSID, BSSID w/ colons, RSSI)
- Cross-compiles clean for all six targets
- `get_active_bssid` logic lives as a testable function, wired into log output
- README/AGENTS.md updated to the new build/install flow
- Bash script removed only after parity confirmed

## Out of scope (follow-on issues)

- iwd/wpa_supplicant backends
- Windows and macOS backends + their event hooks
- roam-detector logic and the adaptive-dropoff feedback loop
