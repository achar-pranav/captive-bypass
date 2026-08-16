# Rewrite: Go + Fyne

> **Historical record.** Superseded by GitHub issues #5-#13. Some decisions here
> (10s poller, "no NM dispatcher", pre-login logout) have since changed — see
> the issues and docs/DESIGN.md for the current design.

Replace the bash script with a Go binary + Fyne GUI. One binary that works on
student laptops (Windows, macOS, Linux), built backend-first so the portal logic
stays transport-agnostic.

## Settled decisions

- **Go** for the core; **Fyne** (pure Go) for the GUI. No web tech, no other languages.
- **One binary, two modes**:
  - `serve` — headless poller (~10s loop): on a trigger SSID and not paused -> login.
  - `gui` — Fyne control panel: setup wizard, status, pause, change/delete creds, logout button, manage SSIDs.
  - The **config file is the contract** between them; they never need to run together.
- **Config**: one `config.json` (0600, in a 0700 dir) with `ssids[]`, `paused`,
  `portal`, `timings`, and creds encrypted with a **machine-fingerprint key**
  (scrypt + AES-GCM). No keyring, no prompts, no plaintext on disk.
- **State** in a separate `state` file (written only by `serve`), so two writers never race.
- **Zero-sudo autostart**: Windows per-user scheduled task, macOS LaunchAgent,
  Linux systemd-user / autostart. No NM dispatcher in the rewrite.
- **Manual logout button in the GUI** is the "I'm leaving" convenience. No
  predictive/adaptive roam logic in v1 (see VANGUARD.md).
- **bash script stays** until the Go binary passes parity, then it's deleted.
- **Linux/NetworkManager (nmcli) is the first backend**; others slot into `backends/`.

## Layout

```
cmd/captive-bypass/   main; subcommands: login, logout, serve, gui, --install, --uninstall, --help
internal/
  portal/   Sophos/Cyberoam login+logout (login.xml mode=191, logout.xml mode=193)
  config/   config.json + fingerprint-encrypted creds
  state/    state file (last login/logout, BSSID)
  serve/    poller loop
  gui/      Fyne app
backends/
  nmcli/    NetworkManager (Linux) — first backend
```

## Preserved from bash (regression baseline)

Login on trigger SSID, pre-login logout + gap, retries + cooldowns, pause
(disabled) flag, env overrides, first-run setup flow.

## Done when

- Go binary reproduces every current bash flow (manual parity run)
- `serve` auto-logs in on campus WiFi, headless
- `gui` wizard + logout button work
- Cross-compiles clean for linux/windows/darwin (amd64+arm64)
- bash script deleted after parity
