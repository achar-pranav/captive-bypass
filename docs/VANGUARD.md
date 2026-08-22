# Vanguard (roam handling)

**Status: confirmed necessary. BSSID-per-classroom makes stale sessions a daily
problem, not an edge case.**

## What it is

Vanguard monitors RSSI trend + BSSID drift while associated and fires a
predictive logout before the radio drops. It also handles re-login when the
signal stabilizes at a weak threshold.

`serve --vanguard` appends one line per poller tick to a **local file**
(`~/.config/captive-bypass/vanguard.log`):

```
<time> <ssid> <bssid> <rssi>
```

The log is for tuning. The actual behavior is the state machine below.

## Confirmed behavior

BSSID changes per classroom (same SSID). Walking class-to-class triggers
"maximum login limit reached" — the portal's stale session was never released
because the AP changed. This happens on every classroom transition, not just
long-distance roams.

## Target flow

The state machine monitors RSSI trend while associated:

```
                 ┌──────────────────────────────────┐
                 │                                  │
                 ▼                                  │
           ┌──────────┐    signal drops      ┌─────────────┐
           │  STRONG  │ ──────────────────▶  │  WEAKENING  │
           └──────────┘                      └─────────────┘
                 ▲                            │            │
                 │                stabilizes  │            │ crosses
                 │             above threshold│            │ weak threshold
                 │                            ▼            ▼
                 │                      ┌──────────┐  ┌──────────────┐
                 │                      │  STRONG  │  │ EDGE-OF-COV  │
                 └──────────────────────┘(recovery)│  │ re-login +   │
                                                  │  │ GUI warning  │
                                                  │  └──────────────┘
                                                  │
                     signal ramps back up ─────────┘
```

**State transitions:**

1. **STRONG → WEAKENING:** RSSI trend goes negative (strong → weak). Fire
   predictive logout while still on old AP (signal still usable, link alive).
2. **WEAKENING → STRONG:** Signal stabilizes above weak threshold. No action —
   connection is fine, false alarm.
3. **WEAKENING → EDGE-OF-COV:** Signal crosses weak threshold (e.g. -75 dBm).
   Auto re-login + GUI notification: "edge of coverage."

**Known risk: rubberbanding** — signal bounces around the threshold, causing
multiple logout/login cycles. Mitigation: trailing-average smoothing + cooldown
between transitions.

## How to collect

1. Enable the flag, walk your route (classroom -> classroom, building -> building).
2. Copy `vanguard.log` back to the maintainer.

**No telemetry, no endpoint, no permissions** — the file never leaves the
machine on its own. Collection is opt-in via the open-source club.

## Tuning targets

Measure from vanguard.log data:

- **RSSI at which BSSID changes** — what dBm does the AP handoff happen at?
- **Seconds of warning** — how long does the RSSI ramp last before radio drop?
- **Weak threshold** — what dBm value triggers edge-of-coverage? (target: -75 dBm)
- **Stabilization window** — how long must RSSI stay above threshold to count
  as "recovered"? (target: 5-10s to avoid rubberbanding)
- **BSSID change timing** — does BSSID flip before or after the RSSI drop?
