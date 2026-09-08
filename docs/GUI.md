# captive-bypass GUI Specification & Design

This document describes the modern GUI architecture for `captive-bypass`, utilizing **Wails v2** (Go + Native OS Webview) with a minimalist **AMOLED true-black** palette, electric blue highlights, responsive anti-spam interactions, and non-intrusive bottom toasts.

---

## 1. Architectural Pivot: Fyne → Wails v2

### Why Fyne was Replaced
- **Rigid Widget Metrics**: Fyne renders via its own OpenGL/Metal canvas primitives with inflexible built-in padding, fixed button shapes, and no CSS support.
- **Styling Clashes**: Achieving custom AMOLED dark themes in Fyne required hacking layout min-sizes and canvas primitives, resulting in an outdated Java/Material aesthetic rather than modern web polish.
- **Widget Duplication**: Password visibility toggles required separate auxiliary button widgets placed beside entries, resulting in duplicated or misaligned visual boxes.

### Why Wails v2
- **Native OS Webview (No Electron Bloat)**: Uses the browser engine already pre-installed on the host OS:
  - **Windows**: Microsoft Edge WebView2 (pre-installed on Windows 10 & 11)
  - **macOS**: Apple WebKit / `WKWebView` (native macOS framework)
  - **Linux**: `libwebkit2gtk-4.0 / 4.1` (standard desktop package)
- **Compact Binary Size**: Stripped executable is **~12 – 15 MB** with ~35MB RAM usage (virtually identical to or lighter than Fyne with custom fonts).
- **100% Pixel-Perfect Styling**: Uses standard HTML, CSS, Tailwind CSS, and React.
- **Embedded Distribution**: Production assets in `internal/gui/frontend/dist` are embedded into the Go binary at compile-time via Go's native `//go:embed`. The main package lives at the repo root (Wails requires it next to `wails.json`), and the frontend must be built with `npm run build` before `go build` so `frontend/dist` is current.
- **Direct Go ↔ TS RPC**: The frontend invokes Go backend methods directly through `window.go.gui.App` with automatic type safety.

---

## 2. Visual Theme & Styling

- **Background Canvas**: Pitch black (`#000000`) for true AMOLED displays.
- **Card Surfaces**: Deep dark slate (`#0A0C0E` and `#0E1012`).
- **Accent Color**: Electric Blue (`#00A8FF`).
- **Primary Action Buttons**: Electric blue background with bold black text, no border outline (`#00A8FF` hover `#33BAFF`).
- **Secondary Buttons**: AMOLED black background with white text and a rounded border outline (`border-white/20 hover:border-white/40`), with a subtle white tint on hover.
- **Delete Buttons**: Compact red square buttons with a trash can icon (`🗑`).
- **Password Input**: Single unified input field with an inside right-aligned eye toggle (`👁` / `🚫`) with zero widget duplication.
- **Bottom Toast Notifications**: Smooth non-intrusive popup that spawns from the bottom of the window. Clicking the popup immediately dismisses it.

---

## 3. Initial Setup Wizard (3 Screens)

Triggered automatically on first run when no credentials are configured, or via setup reset.

### Screen 1: Permissions & Trust
- Informs the user why Wi-Fi interface observation is needed.
- Establishes program trust: 100% open-source, local-only execution, zero root permissions required.
- **Continue**: Advances to Screen 2.
- **Skip**: Jumps directly to the Main Menu.

### Screen 2: Add Credentials
- 3 primary input fields:
  1. **Credentials name** (default: `default`)
  2. **Username** (SRN, e.g. `PES1UG...`)
  3. **Password** (Portal password with inside eye toggle button to show/hide typed password)
- **Security Disclaimer**:
  > *"Passwords are never stored in plaintext. We use OS hardware fingerprinting and AES-GCM encryption to prevent theft by copy."*
- **Back**: Secondary button placed above Continue to return to Screen 1.
- **Continue**: Full-width electric blue button at the bottom. Validates inputs, encrypts credentials, and advances to Screen 3.

### Screen 3: Select at least one SSID to log into automatically
- Header title: *"Select at least one SSID to log into automatically"*
- **Networks ⟳** refresh button next to the title to re-scan surrounding APs.
- Wi-Fi list with checkboxes and live signal percentage (`%`).
- **Back**: Secondary button placed above Done to return to Screen 2.
- **Done**: Full-width electric blue button at the bottom. Saves selected SSIDs and launches the Main Menu.

---

## 4. Main Menu

A clean, minimalist control panel with no superfluous logs or clutter.

### Status Header & 4-State LED
Located at the top left of the interface:
- **🟢 Green**: `<Connected>` (Subtext: `Connected to <SSID>`) — Wi-Fi connected and portal authenticated.
- **🔴 Red**: `<Disconnected>` (Subtext: `Wi-Fi is offline` or `captive-bypass paused`) — Wi-Fi offline or bypass disabled.
- **🟡 Yellow**: `<In Progress>` (Subtext: `Authenticating with <SSID>`) — Associated with target Wi-Fi, negotiating captive portal login handshake.
- **🟠 Orange**: `Network edge` (Subtext: `You're at the edge`) — Triggered when signal drops below configured threshold (default 15%). No text wrapping.

### Networks Card
- Displays count of recognized networks.
- Two action buttons on the far right:
  - **[Add]**: Opens the network picker dialog with checkboxes and `Networks ⟳` button to register new SSIDs.
  - **[Manage]**: Opens a list of all registered networks with red square trash buttons (`🗑`) to delete entries (even when offline).

### Creds Card
- Displays active profile name and configured sets count.
- Two action buttons on the far right:
  - **[Add]**: Opens the 3-field credential entry dialog (Name, Username, Password with eye toggle) with encryption disclaimer.
  - **[Manage]**: Opens the profile management dialog with radio buttons to switch the active set, an **Edit pencil button (`✏️`)** to update credentials/passwords, and trash buttons (`🗑`) to delete profiles.

### Bottom Preference Sliders
1. `[x] Enable/Disable the captive-bypass` (pauses or resumes auto-login).
2. `[x] Enable/Disable Vanguard(experimental)` (toggles edge-of-network telemetry and proactive session retention).

---

## 5. Development & Compilation

### Standard Go Build (Production)
```bash
wails build
./build/bin/captive-bypass gui
```

Note: `go build` requires the frontend to already be built (`npm run build` in `internal/gui/frontend`) so the embedded `frontend/dist` assets exist. `wails build` handles that step automatically.

### Wails Development (with live reload)
```bash
wails dev
```
