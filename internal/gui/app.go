package gui

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/achar-pranav/captive-bypass/backends"
	"github.com/achar-pranav/captive-bypass/internal/config"
	"github.com/achar-pranav/captive-bypass/internal/portal"
)

// App is the main Wails application bridge exposing Go backend methods to the frontend.
type App struct {
	ctx     context.Context
	cfg     *config.Config
	cfgPath string
	portal  *portal.Client
	wifi    backends.Backend
	mu      sync.Mutex
}

// CredProfileInfo represents a credential profile summary without exposing secrets.
type CredProfileInfo struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	IsActive bool   `json:"isActive"`
}

// AppState represents the complete state of the captive portal daemon and Wi-Fi interface.
type AppState struct {
	Status             string            `json:"status"` // "green" | "red" | "yellow" | "orange"
	StatusTitle        string            `json:"statusTitle"`
	StatusSub          string            `json:"statusSub"`
	ActiveSSID         string            `json:"activeSSID"`
	SignalPercent      int               `json:"signalPercent"`
	IsAutoLoginEnabled bool              `json:"isAutoLoginEnabled"`
	IsVanguardEnabled  bool              `json:"isVanguardEnabled"`
	Threshold          int               `json:"threshold"`
	ActiveProfile      string            `json:"activeProfile"`
	CredsCount         int               `json:"credsCount"`
	RecognizedNetworks []string          `json:"recognizedNetworks"`
	CredProfiles       []CredProfileInfo `json:"credProfiles"`
	IsFirstRun         bool              `json:"isFirstRun"`
}

// ScannedNetwork represents a nearby Wi-Fi network.
type ScannedNetwork struct {
	SSID    string `json:"ssid"`
	Signal  int    `json:"signal"`
	Secured bool   `json:"secured"`
}

// NewApp creates a new App bridge instance.
func NewApp(cfg *config.Config, cfgPath string, pc *portal.Client, b backends.Backend) *App {
	return &App{
		cfg:     cfg,
		cfgPath: cfgPath,
		portal:  pc,
		wifi:    b,
	}
}

// Startup is called when the Wails application starts up.
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

// GetState returns the current application, network, and profile state.
func (a *App) GetState() AppState {
	a.mu.Lock()
	defer a.mu.Unlock()

	ssid := ""
	signal := 0
	up := false

	if a.wifi != nil {
		if s, err := a.wifi.ActiveSSID(); err == nil {
			ssid = s
		}
		if sig, err := a.wifi.Signal(); err == nil {
			signal = dbmToPercent(sig)
		}
		if u, err := a.wifi.Up(); err == nil {
			up = u
		}
	}

	status := "red"
	statusTitle := "<Disconnected>"
	statusSub := "Wi-Fi is offline"

	threshold := a.cfg.SignalThreshold()

	if a.cfg.Paused {
		status = "red"
		statusTitle = "<Disabled>"
		statusSub = "captive-bypass paused"
	} else if !up || ssid == "" {
		status = "red"
		statusTitle = "<Disconnected>"
		statusSub = "Wi-Fi is offline"
	} else if signal > 0 && signal <= threshold {
		status = "orange"
		statusTitle = "Network edge"
		statusSub = "You're at the edge"
	} else {
		// Check if active SSID is registered
		isRegistered := false
		for _, reg := range a.cfg.SSIDs {
			if reg == ssid {
				isRegistered = true
				break
			}
		}

		if isRegistered {
			// Check live portal status
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			online, _ := a.portal.Livecheck(ctx)
			cancel()

			if online {
				status = "green"
				statusTitle = "<Connected>"
				statusSub = fmt.Sprintf("Connected to %s", ssid)
			} else {
				status = "yellow"
				statusTitle = "<In Progress>"
				statusSub = fmt.Sprintf("Authenticating with %s", ssid)
			}
		} else {
			status = "red"
			statusTitle = "<Disconnected>"
			statusSub = fmt.Sprintf("Connected to %s (not registered)", ssid)
		}
	}

	profiles := make([]CredProfileInfo, 0, len(a.cfg.CredSets))
	for _, cs := range a.cfg.CredSets {
		profiles = append(profiles, CredProfileInfo{
			Name:     cs.Name,
			Username: cs.Username,
			IsActive: cs.Name == a.cfg.ActiveSet,
		})
	}

	return AppState{
		Status:             status,
		StatusTitle:        statusTitle,
		StatusSub:          statusSub,
		ActiveSSID:         ssid,
		SignalPercent:      signal,
		IsAutoLoginEnabled: !a.cfg.Paused,
		IsVanguardEnabled:  a.cfg.Vanguard,
		Threshold:          threshold,
		ActiveProfile:      a.cfg.ActiveSet,
		CredsCount:         len(a.cfg.CredSets),
		RecognizedNetworks: a.cfg.SSIDs,
		CredProfiles:       profiles,
		IsFirstRun:         len(a.cfg.CredSets) == 0,
	}
}

// ScanNetworks performs a Wi-Fi scan and returns deduplicated SSIDs sorted by signal strength.
func (a *App) ScanNetworks() ([]ScannedNetwork, error) {
	if a.wifi == nil {
		return nil, fmt.Errorf("no wifi backend available")
	}

	aps, err := a.wifi.Scan()
	if err != nil {
		return nil, err
	}

	consolidated := backends.Consolidate(aps)
	out := make([]ScannedNetwork, 0, len(consolidated))
	for _, ap := range consolidated {
		sig := ap.Signal
		if sig < 0 {
			sig = dbmToPercent(sig)
		}
		out = append(out, ScannedNetwork{
			SSID:    ap.SSID,
			Signal:  sig,
			Secured: ap.Secured,
		})
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].Signal > out[j].Signal
	})

	return out, nil
}

// AddSSID registers an SSID to the auto-login recognized list.
func (a *App) AddSSID(ssid string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	for _, s := range a.cfg.SSIDs {
		if s == ssid {
			return nil
		}
	}

	a.cfg.SSIDs = append(a.cfg.SSIDs, ssid)
	return config.Save(a.cfgPath, a.cfg)
}

// RemoveSSID unregisters an SSID from the recognized list.
func (a *App) RemoveSSID(ssid string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	filtered := make([]string, 0, len(a.cfg.SSIDs))
	for _, s := range a.cfg.SSIDs {
		if s != ssid {
			filtered = append(filtered, s)
		}
	}

	a.cfg.SSIDs = filtered
	return config.Save(a.cfgPath, a.cfg)
}

// SaveCreds encrypts and saves credentials for a named profile.
func (a *App) SaveCreds(name, username, password string, setActive bool) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if name == "" {
		name = "default"
	}

	fp, err := config.MachineFingerprint()
	if err != nil {
		return fmt.Errorf("deriving hardware fingerprint: %w", err)
	}

	if err := a.cfg.SetCredSet(fp, name, username, password); err != nil {
		return fmt.Errorf("encrypting credentials: %w", err)
	}

	if setActive || a.cfg.ActiveSet == "" {
		a.cfg.ActiveSet = name
	}

	return config.Save(a.cfgPath, a.cfg)
}

// GetCreds decrypts and retrieves a credential profile for editing.
func (a *App) GetCreds(name string) (map[string]string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	fp, err := config.MachineFingerprint()
	if err != nil {
		return nil, fmt.Errorf("deriving hardware fingerprint: %w", err)
	}

	user, pass, err := a.cfg.GetCredsByName(fp, name)
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"name":     name,
		"username": user,
		"password": pass,
	}, nil
}

// DeleteCreds removes a credential profile by name.
func (a *App) DeleteCreds(name string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if err := a.cfg.DeleteCredSet(name); err != nil {
		return err
	}

	if a.cfg.ActiveSet == name || a.cfg.ActiveSet == "" {
		if len(a.cfg.CredSets) > 0 {
			a.cfg.ActiveSet = a.cfg.CredSets[0].Name
		} else {
			a.cfg.ActiveSet = ""
		}
	}

	return config.Save(a.cfgPath, a.cfg)
}

// SetActiveCred switches the active credential profile.
func (a *App) SetActiveCred(name string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if err := a.cfg.SetActiveSet(name); err != nil {
		return err
	}

	return config.Save(a.cfgPath, a.cfg)
}

// ToggleAutoLogin toggles the master auto-login switch.
func (a *App) ToggleAutoLogin(enabled bool) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.cfg.Paused = !enabled
	return config.Save(a.cfgPath, a.cfg)
}

// ToggleVanguard toggles the Vanguard edge-of-network telemetry mode.
func (a *App) ToggleVanguard(enabled bool) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.cfg.Vanguard = enabled
	return config.Save(a.cfgPath, a.cfg)
}

// SetThreshold updates the edge-of-network signal percentage threshold.
func (a *App) SetThreshold(threshold int) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if threshold < 1 || threshold > 100 {
		threshold = 15
	}
	a.cfg.Threshold = threshold
	return config.Save(a.cfgPath, a.cfg)
}

// ManualLogin triggers an explicit login attempt against the captive portal.
func (a *App) ManualLogin() (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	fp, err := config.MachineFingerprint()
	if err != nil {
		return "", fmt.Errorf("hardware fingerprint: %w", err)
	}

	user, pass, err := a.cfg.GetActiveCreds(fp)
	if err != nil {
		return "", fmt.Errorf("reading active credentials: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ok, msg, err := a.portal.Login(ctx, user, pass)
	if err != nil {
		return "", err
	}
	if !ok {
		return fmt.Sprintf("Portal login failed: %s", msg), nil
	}

	return "Logged in successfully", nil
}

// ManualLogout triggers an explicit logout attempt against the captive portal.
func (a *App) ManualLogout() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	user, err := a.cfg.ActiveUser()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return a.portal.Logout(ctx, user)
}

// FinishWizard completes the setup wizard by committing initial SSIDs.
func (a *App) FinishWizard(ssids []string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.cfg.SSIDs = ssids
	return config.Save(a.cfgPath, a.cfg)
}

func dbmToPercent(dbm int) int {
	if dbm >= -50 {
		return 100
	}
	if dbm <= -100 {
		return 0
	}
	return 2 * (dbm + 100)
}
