# captive-bypass

Auto-login for the PESU **Sophos/Cyberoam** captive portal (`rr.pes.edu:8090`), so you never have to open a browser on the ELEMENT BLOCK WiFi.

**Status:** the tool is being rewritten as a **Go binary** (CLI + Fyne GUI) targeting Linux, Windows, and macOS. The plan lives in the GitHub issues. The original bash tool is archived at `attic/captive-bypass` and used only as a protocol reference.

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

## Backends

The design is **backend-tiered**: WiFi managers (nmcli first, then iwd/wpa_supplicant) and OSes (Linux, then Windows/macOS) are additional backends rather than forks — the portal logic is transport-agnostic.

## References

- `attic/captive-bypass` — archived bash tool (reference only)
- `docs/GO-REWRITE.md` — historical rewrite RFC (superseded by issues)
- `docs/VANGUARD.md` — experimental data-collection design (low priority)
- `docs/DESIGN.md` — design notes + ELI5 glossary
- GitHub issues — the current plan
