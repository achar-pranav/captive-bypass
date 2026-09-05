# captive-bypass

Auto-login for the PESU **Sophos/Cyberoam** captive portal (`rr.pes.edu:8090`), so you never have to open a browser on the ELEMENT BLOCK WiFi.

**Status:** Linux is working end-to-end (verified live): kernel-event watcher → auto login/logout, zero privilege prompts. Windows backend is written, awaiting a real machine to test. macOS CoreWLAN backend implemented (#23). The plan lives in the GitHub issues; the original bash tool is archived at `attic/captive-bypass` (protocol reference only).

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

## Install & use (Linux)

```
go build -o captive-bypass ./cmd/captive-bypass
./captive-bypass --install      # systemd user unit; no password prompts, ever
./captive-bypass gui            # first-run wizard: cred set + SSID picker
```

CLI mirrors the old bash tool's verbs:

| verb | what it does |
|---|---|
| `login` / `logout` | portal login (best-effort logout first) / logout |
| `enable` / `disable` | master auto-login switch |
| `update-creds --user SRN --pass PW [--name N]` | add/update a named credential set |
| `set-network SSID...` | register SSIDs that trigger login |
| `serve` | run the watcher daemon (the unit runs this) |
| `gui` | control panel |
| `dev wipe` / `reset-state` / `clear-vanguard` / `force` | tester helpers |

## How auto-login works

`serve` subscribes to **kernel netlink events** (link/address/route) as your
plain user — it listens below any WiFi manager, so NetworkManager, iwd,
wpa_supplicant or none all look identical. Events are debounced (~0.8 s), the
connection state is read via `iw`, and if the SSID is one you registered and
you're not already online, it logs in; on disconnect it best-efforts a logout
so roaming doesn't hit the portal's session limit. No heartbeat, no root.

## Config

Everything lives in `~/.config/captive-bypass/`: encrypted credential sets +
which is active (`config.json`), recognized SSIDs, session state, log.
Credentials are AES-GCM-encrypted with a machine-derived key — they never
leave the device except in the portal's own login POST (which the protocol
sends plaintext; see above).

## Backends

The design is backend-tiered: the kernel watcher + `iw` state reader cover all
Linux managers; Windows uses its WLAN notification API; macOS uses CoreWLAN via CGO (#23). The portal logic is transport-agnostic.

## References

- `attic/captive-bypass` — archived bash tool (reference only)
- `docs/GO-REWRITE.md` — historical rewrite RFC (superseded by issues)
- `docs/VANGUARD.md` — experimental data-collection design (low priority)
- `docs/DESIGN.md` — design notes + ELI5 glossary
- GitHub issues — the current plan
