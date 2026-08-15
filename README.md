# captive-bypass

Auto-login for the PESU **Sophos/Cyberoam** captive portal (`rr.pes.edu:8090`), so you never have to open a browser on the ELEMENT BLOCK WiFi.

Written in **bash** on purpose: the login is a single stateless POST, and the response is trivially parsed — Python would add a runtime dependency without buying anything cleaner. The tool needs only `curl`, `nmcli` (NetworkManager), and `sudo`, all near-universal on Linux.

**For agents/AI sessions:** read `AGENTS.md` first (auto-loaded); design context and the ELI5 glossary live in `docs/DESIGN.md`; tasks are tracked as GitHub issues.

## How the portal login works

Reverse-engineered from the portal's own `httpclient.js`:

```
POST https://rr.pes.edu:8090/login.xml
Content-Type: application/x-www-form-urlencoded

mode=191&username=<SRN>&password=<PASSWORD>&a=<epoch-ms>&producttype=0
```

- `mode=191` fixed (login); `193` = logout.
- `password` is sent **plaintext**, URL-encoded. No client-side hashing.
- `a` = current epoch time in **milliseconds**.
- No cookies, no CSRF token, no session setup. Stateless.
- Success = response body contains `<status><![CDATA[LIVE]]></status>`.
- Failure = `<status><![CDATA[LOGIN]]></status>` with a reason in `<message>`.

## Requirements

- Linux with **NetworkManager** (`nmcli`), `curl`, `sudo`
- bash

`iwd` / `wpa_supplicant` are **not** supported yet — see [Known limitations](#known-limitations).

## Install & first run

```sh
git clone <your-repo> && cd captive-bypass   # or just copy this directory
./captive-bypass
```

First run prompts you for:

1. **SRN** and **password** (portal credentials)
2. The **trigger SSID** — it detects the currently connected WiFi, shows it to you, and asks for confirmation (`[y/N]`) before saving anything
3. Whether to install the NetworkManager dispatcher (`[y/N]`, runs one `sudo` write)

After setup it immediately tries to log in. Nothing is ever assumed silently: the SSID is only saved after explicit confirmation, and credentials are only re-prompted when you ask for it (`--update-creds`).

### Installing the auto-login dispatcher

First-run setup asks, but you can install (or reinstall) later:

```sh
./captive-bypass --install
```

This writes exactly **one root-owned file**: `/etc/NetworkManager/dispatcher.d/90-captive-bypass`. Everything else lives in your user config. The dispatcher fires `captive-bypass login` (backgrounded, with a 3s settle delay) whenever any interface comes up; the tool itself re-checks that the active SSID matches before doing anything, so nothing fires on the wrong network.

## Usage

| Command | What it does |
|---|---|
| `captive-bypass` | First-run setup (creds + SSID capture), or just log in if already configured |
| `captive-bypass login` | Force a login attempt now (only if active SSID matches the stored one) |
| `captive-bypass --update-creds` | Re-enter SRN/password and overwrite the stored file |
| `captive-bypass --set-network` | Re-capture the trigger SSID |
| `captive-bypass --install` | Install the NM dispatcher (needs sudo) |
| `captive-bypass --uninstall` | Remove the dispatcher (needs sudo) and delete the config directory |
| `captive-bypass --help` | Show usage |

### Retry behavior

On a failed login the tool retries **3 times maximum** with a 5s delay between attempts, then stops with:

```
Login failed 3x - check credentials with 'captive-bypass --update-creds', or the portal may be down
```

It never loops indefinitely.

## Configuration & logs

- Credentials: `~/.config/captive-bypass/creds`
- Trigger SSID: `~/.config/captive-bypass/network.conf`
- Log: `~/.config/captive-bypass/log`

## Security model: why credentials are stored plaintext (chmod 600), not openssl-encrypted

The original design goal was `openssl enc` AES at rest, with a passphrase prompted once at save and again on each run. **That conflicts with the core purpose of this tool.** Auto-login on connect runs from the NM dispatcher: no TTY, no interactive context. A per-run passphrase prompt would make the tool fail (or hang) every time an interface comes up — exactly the unattended scenario it exists for.

There is also no way to get encryption for free: any passphrase/key that a user-space script can read unattended is itself machine-readable, so encrypting the file would only move the secret next to the secret. On the threat model that matters here — casual inspection, accidental sharing of the config dir, `git` slip-ups — `chmod 600` on a file inside a `chmod 700` directory provides the same practical protection.

**Tradeoff (chosen):** credentials are stored in plaintext at `~/.config/captive-bypass/creds`, mode `0600`, inside a mode `0700` directory, owned by you. A same-user compromise (or root) defeats any of these options equally.

If you genuinely want encryption at rest, the honest mechanism is an **OS keyring** (GNOME Keyring / KWallet / `pass`), which holds the unlock secret in a trusted, kernel-protected store instead of the filesystem — see [Future work](#future-work). `openssl enc` was rejected, not because it is hard, but because it cannot be both encrypted-at-rest and decryptable unattended without being equivalent to plaintext.

## Known limitations

- **NetworkManager only.** SSID detection relies on `nmcli`; `iwd` and `wpa_supplicant` are not handled. Adding them means teaching `get_active_ssid()` to speak their CLIs (`iwctl`, `wpa_cli`) — the rest of the tool is transport-agnostic.
- The dispatcher only fires on interface `up`. If the portal drops you mid-session, nothing re-logs you in (see keep-alive below).

## Future work

- **Session keep-alive / polling** for mid-session logout (the portal can silently drop long-lived sessions).
- **Multi-SSID support** — currently exactly one trigger network.
- **CAPTCHA / CHALLENGE-state handling** — the portal has a `CHALLENGE` status with a `state` value to echo back; unimplemented.
- **OS keyring integration** for encrypted-at-rest credentials that still unlocks unattended (system keyring, not `openssl enc`).
- **LICENSE file** — intentionally not added yet.

## Troubleshooting

- **"Not on trigger network ... skipping login"** — you're on a different SSID. Re-run `./captive-bypass --set-network` while connected to the network you want to trigger on.
- **Repeated login failures** — re-enter credentials with `./captive-bypass --update-creds`, or the portal may be down.
- **nmcli not found** — NetworkManager isn't installed or `nmcli` isn't in the dispatcher's PATH; install NetworkManager.
- **Debugging** — check `~/.config/captive-bypass/log`; every attempt and its failure reason is timestamped there. Test against a mock portal with `CAPTIVE_BYPASS_PORTAL=http://127.0.0.1:PORT` and shorten retries with `CAPTIVE_BYPASS_RETRY_DELAY=1`.
