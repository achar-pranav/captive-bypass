# AGENTS.md — captive-bypass

## What this is
A tool that auto-logs into the PESU Sophos/Cyberoam captive portal
(rr.pes.edu:8090) using stored SRN/password, triggered on campus WiFi connect.
Current tool: **Go binary** with tiered backends. The old bash script is
archived in `attic/` and used only as a protocol reference. Design is
backend-tiered to cover all WiFi managers and OSes (student laptops), so treat
OS/manager specifics as swappable backends, not permanent facts.

## Platform priority
Windows and macOS are first-class until v1 is out — not afterthoughts. Linux is
the dev machine, but every design decision must keep Windows/macOS buildable
and testable. Do not write code that only works on Linux unless an issue says
so.

## The one non-negotiable fact (roaming)
The portal keys a session to MAC + network location (AP/switch). When the user
roams to another AP (same SSID, e.g. floor 1->4), the old session is never
released, so a new login is rejected with "maximum login limit reached". A
logout (POST logout.xml mode=193) ONLY releases a session if sent from the SAME
location as the login; sending it from the new AP does nothing.

Consequence: roam-handling must fire logout BEFORE association to the old AP is
lost (signal-strength trend + BSSID delta, while still in range). Hard radio
drops are not recoverable client-side; worst case is lockout until the portal's
session timeout (~3h, unmeasured). MAC randomization per connect is the only
100% reliable sidestep, but carries policy risk (looks like device churn to IT).

## Files
- `backends/` — Backend interface + implementations:
  - `nmcli/` — NetworkManager connectivity check + SSID scan
  - `iw/` — `iw` state reader (SSID, signal, BSSID for watcher)
  - `windows/` — WLAN API backend (events, listener, netsh parsing)
  - `auto/` — platform-aware auto-select of the right backend
- `cmd/captive-bypass/main.go` — CLI entry point
- `internal/` — core logic:
  - `config/` — config.json + AES-GCM fingerprint-encrypted cred sets
  - `state/` — cooldowns, last BSSID, session state
  - `portal/` — login/logout over HTTPS (mode 191/193)
  - `serve/` — watcher daemon (unix socket, event-driven)
  - `watcher/` — kernel netlink event subscriber (Linux)
  - `gui/` — Fyne setup wizard + control panel
  - `install/` — systemd user unit install/uninstall
- `attic/captive-bypass` — archived bash tool (protocol reference only)
- `attic/dispatcher-hook.sh` — old dispatcher hook (reference only)
- `docs/`:
  - `DESIGN.md` — design notes + ELI5 glossary; read ONLY when an issue references it
  - `GO-REWRITE.md` — historical rewrite RFC (superseded by GitHub issues)
  - `VANGUARD.md` — experimental data-collection design (low priority)
- `README.md` — install/usage/security

## Config (created at runtime)
- `~/.config/captive-bypass/` — `config.json`, cred sets (encrypted), `state.json`, log

## Workflow contract (how we work)
- One GitHub issue per task; the issue body is the only briefing needed.
- Every session starts with: this file + `gh issue view <N>` for the target.
  ALL context comes from these two; nothing else is fetched unless an issue
  says so (e.g. "see docs/DESIGN.md §roaming").
- Patch only code named in the issue's Files section; reference functions, not
  line numbers. Keep the change byte-sized.
- Verify against the issue's "Done when" checklist. No test suite; manual verify.
- Never commit, create issues, or push unless the user explicitly says so.
- If a session ends mid-issue, update the issue's checklist first so the next
  session resumes cleanly.
- The user is not a strong dev: answer architecture questions in short
  explanations; never dump whole files into chat.
- When the user asks to ELI5 any concept, coding practice, or framework,
  assume they are a non-coder: explain in plain English with practical
  "what does this mean for us" framing — never explain like a developer
  would to another developer.
- Whenever a framework/approach is picked (GUI, backend, storage, etc.), add an
  entry to docs/DESIGN.md listing the alternatives we looked at and why each
  was NOT chosen, before moving on.
- Walk the dev through EVERY step and decision, down to the smallest divisible
  decision, before writing code. The dev decides what and why; the model only
  handles syntax. Never make a design/behavior decision on the dev's behalf.

## Tooling
- `gh`, `go`; `godbus` (Fyne dep), `golang.org/x/sys` (netlink).
- Linux backends: `nmcli` + `iw` (both read state below the WiFi manager).
- Windows backend: `netsh` + WLAN notification API.
- GUI: Fyne v2 (cross-platform).
- No test suite; verification is manual/empirical.