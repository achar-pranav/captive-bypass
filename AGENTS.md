# AGENTS.md — captive-bypass

## What this is
A bash tool that auto-logs into the PESU Sophos/Cyberoam captive portal
(rr.pes.edu:8090) using stored SRN/password, triggered on campus WiFi connect.

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
- captive-bypass     the whole tool (single bash script)
- docs/DESIGN.md     design notes + ELI5 glossary (human-facing, fetched on demand)
- README.md          install/usage/security

## Config (created at runtime)
- ~/.config/captive-bypass/  creds, network.conf, state, log, disabled

## Workflow contract (how we work)
- One GitHub issue per task; the issue body is the only briefing needed.
- Every session starts with: this file + `gh issue view <N>` for the target.
- Patch only code named in the issue's Files section; reference functions, not
  line numbers. Keep the change byte-sized.
- Verify against the issue's "Done when" checklist. No test suite; manual verify.
- Never commit, create issues, or push unless the user explicitly says so.
- If a session ends mid-issue, update the issue's checklist first so the next
  session resumes cleanly.
- The user is not a strong dev: answer architecture questions in short
  explanations; never dump whole files into chat.

## Tooling
- gh (auth achar-pranav), nmcli (NetworkManager), curl, bash. Linux-only.
