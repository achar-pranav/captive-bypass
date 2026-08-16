# Vanguard (experimental)

**Status: experimental flag, no behavior change. May be dead code if IT ever
fixes the portal's session table to follow switches server-side.**

## What it is

`serve --vanguard` appends one line per poller tick to a **local file**
(`~/.config/captive-bypass/vanguard.log`):

```
<time> <ssid> <bssid> <rssi>
```

That's it. No logic, no adaptive anything — just data.

## Why

The campus portal ties a login to MAC + AP location. Roaming to another AP
leaves a stale session behind ("maximum login limit reached"). The eventual fix
(a pre-emptive logout fired while signal is still usable) needs real walk data
to tune against. This flag collects that data.

## How to collect

1. Enable the flag, walk your route (classroom -> classroom, building -> building).
2. Copy `vanguard.log` back to the maintainer.

**No telemetry, no endpoint, no permissions** — the file never leaves the
machine on its own. Collection is opt-in via the open-source club.

## Tuning targets (future)

- How many seconds of warning an RSSI ramp gives before the radio dies
- Whether BSSID changes before or after the drop
- A simple threshold/trailing-drops rule (kept deliberately dumb, not "smart")
