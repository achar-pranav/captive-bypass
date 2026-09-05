package gui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/achar-pranav/captive-bypass/backends"
	"github.com/achar-pranav/captive-bypass/backends/auto"
	"github.com/achar-pranav/captive-bypass/internal/config"
	"github.com/achar-pranav/captive-bypass/internal/install"
	"github.com/achar-pranav/captive-bypass/internal/portal"
)

type ui struct {
	a             fyne.App
	w             fyne.Window
	cfg           *config.Config
	cfgDir        string
	portal        *portal.Client
	wifi          backends.Backend
	log           *widget.Entry
	statusLine    *widget.Label
	portalBadge   *widget.Label
	refreshCancel context.CancelFunc
}

func Run() error {
	dir := config.DefaultDir()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	cfgPath := filepath.Join(dir, "config.json")
	cfg, err := config.Load(cfgPath)
	if err != nil && err != config.ErrNoConfig {
		return err
	}
	firstRun := len(cfg.CredSets) == 0

	u := &ui{
		a:      app.NewWithID("io.github.achar-pranav.captive-bypass"),
		cfg:    cfg,
		cfgDir: dir,
		portal: portal.New(cfg.Portal, nil),
		wifi:   auto.Default(),
	}

	u.a.Settings().SetTheme(newAmoledTheme())
	u.w = u.a.NewWindow("captive-bypass")
	u.w.Resize(fyne.NewSize(450, 680))

	if firstRun {
		u.showWizard()
	} else {
		u.showMain()
	}

	u.w.ShowAndRun()
	return nil
}

func (u *ui) saveConfig() {
	if err := config.Save(filepath.Join(u.cfgDir, "config.json"), u.cfg); err != nil {
		u.toast("Could not save settings: " + err.Error())
	}
}

func (u *ui) toast(msg string) {
	dialog.NewInformation("captive-bypass", msg, u.w).Show()
}

// ---------- main window ----------

func (u *ui) showMain() {
	u.statusLine = widget.NewLabel("Checking Wi-Fi connection…")
	u.statusLine.TextStyle = fyne.TextStyle{Bold: true}

	u.portalBadge = widget.NewLabel("Checking portal status…")

	u.log = widget.NewMultiLineEntry()
	u.log.Disable()
	u.log.Wrapping = fyne.TextWrapBreak
	u.log.TextStyle = fyne.TextStyle{Monospace: true}

	statusCard := u.cardStatus()
	credsCard := u.cardCredentials()
	netsCard := u.cardNetworks()
	behaviorCard := u.cardBehavior()
	watcherCard := u.cardWatcher()

	activityBox := container.NewVScroll(u.log)
	activityBox.SetMinSize(fyne.NewSize(0, 140))
	activityCard := widget.NewCard("Activity Log", "", activityBox)

	body := container.NewVBox(
		statusCard,
		credsCard,
		netsCard,
		behaviorCard,
		watcherCard,
	)

	u.w.SetContent(container.NewBorder(nil, activityCard, nil, nil, container.NewScroll(body)))

	if u.refreshCancel != nil {
		u.refreshCancel()
	}
	ctx, cancel := context.WithCancel(context.Background())
	u.refreshCancel = cancel

	u.refreshStatusOnce()
	go u.refreshStatusLoop(ctx)
	go u.refreshLogLoop(ctx)
}

func (u *ui) cardStatus() fyne.CanvasObject {
	signInBtn := widget.NewButton("Sign In Now", u.doSignIn)
	signInBtn.Importance = widget.HighImportance

	signOutBtn := widget.NewButton("Sign Out", u.doSignOut)
	refreshBtn := widget.NewButton("Check Status", u.refreshStatusOnce)

	actions := container.NewHBox(signInBtn, signOutBtn, refreshBtn)
	content := container.NewVBox(
		u.statusLine,
		u.portalBadge,
		actions,
	)
	return widget.NewCard("Status", "", content)
}

func (u *ui) cardCredentials() fyne.CanvasObject {
	names := []string{}
	for _, cs := range u.cfg.CredSets {
		names = append(names, cs.Name)
	}

	radio := widget.NewRadioGroup(names, func(chosen string) {
		if chosen == "" {
			return
		}
		if err := u.cfg.SetActiveSet(chosen); err != nil {
			u.toast(err.Error())
			return
		}
		u.saveConfig()
	})
	radio.SetSelected(u.cfg.ActiveSet)

	addBtn := widget.NewButton("Add / Edit Set", u.showCredsDialog)
	delBtn := widget.NewButton("Delete", func() {
		if u.cfg.ActiveSet == "" || radio.Selected == "" {
			u.toast("Pick a set to delete")
			return
		}
		if err := u.cfg.DeleteCredSet(radio.Selected); err != nil {
			u.toast(err.Error())
			return
		}
		u.saveConfig()
		u.toast("Deleted " + radio.Selected)
		u.showMain()
	})

	hint := widget.NewLabel("The selected credential set is used for portal authentication.")
	hint.TextStyle = fyne.TextStyle{Italic: true}

	var body fyne.CanvasObject
	if len(names) == 0 {
		body = container.NewVBox(
			widget.NewLabel("No credentials stored. Add one to enable auto-login."),
			addBtn,
		)
	} else {
		body = container.NewVBox(
			radio,
			hint,
			container.NewHBox(addBtn, delBtn),
		)
	}
	return widget.NewCard("Credentials", "", body)
}

func (u *ui) cardNetworks() fyne.CanvasObject {
	list := container.NewVBox()
	for _, s := range u.cfg.SSIDs {
		list.Add(widget.NewLabel("• " + s))
	}
	if len(u.cfg.SSIDs) == 0 {
		list.Add(widget.NewLabel("No networks registered yet."))
	}

	manageBtn := widget.NewButton("Scan & Manage Networks", func() { u.showSSIDEditor() })
	body := container.NewVBox(list, manageBtn)
	return widget.NewCard("Recognized Networks", "", body)
}

func (u *ui) cardBehavior() fyne.CanvasObject {
	autoLoginCheck := widget.NewCheck("Auto-login when recognized Wi-Fi connects", func(on bool) {
		u.cfg.Paused = !on
		u.saveConfig()
	})
	autoLoginCheck.SetChecked(!u.cfg.Paused)

	vanguardCheck := widget.NewCheck("Vanguard telemetry (experimental)", func(on bool) {
		u.cfg.Vanguard = on
		u.saveConfig()
	})
	vanguardCheck.SetChecked(u.cfg.Vanguard)

	body := container.NewVBox(autoLoginCheck, vanguardCheck)
	return widget.NewCard("Preferences", "", body)
}

func (u *ui) cardWatcher() fyne.CanvasObject {
	watcherStatus := widget.NewLabel("Checking background service…")

	enableBtn := widget.NewButton("Enable Service", func() {
		go func() {
			err := install.Enable()
			msg := "Auto-start enabled — runs silently in background"
			if err != nil {
				msg = "Enable failed: " + err.Error()
			}
			fyne.Do(func() { u.toast(msg) })
		}()
	})

	disableBtn := widget.NewButton("Disable Service", func() {
		go func() {
			err := install.Disable()
			msg := "Auto-start disabled"
			if err != nil {
				msg = "Disable failed: " + err.Error()
			}
			fyne.Do(func() { u.toast(msg) })
		}()
	})

	go func() {
		for {
			on, _ := install.Status()
			txt := "Background service: Inactive"
			if on {
				txt = "Background service: Active (Starts on login)"
			}
			if watcherStatus.Text != txt {
				fyne.Do(func() { watcherStatus.SetText(txt) })
			}
			time.Sleep(5 * time.Second)
		}
	}()

	body := container.NewVBox(
		widget.NewLabel("Run captive-bypass as a background daemon on boot."),
		container.NewHBox(watcherStatus),
		container.NewHBox(enableBtn, disableBtn),
	)
	return widget.NewCard("Background Daemon", "", body)
}

func (u *ui) refreshStatusOnce() {
	go func() {
		line, badge := u.statusInfo()
		fyne.Do(func() {
			if u.statusLine != nil && u.statusLine.Text != line {
				u.statusLine.SetText(line)
			}
			if u.portalBadge != nil && u.portalBadge.Text != badge {
				u.portalBadge.SetText(badge)
			}
		})
	}()
}

func (u *ui) refreshStatusLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(3 * time.Second):
			line, badge := u.statusInfo()
			fyne.Do(func() {
				if u.statusLine != nil && u.statusLine.Text != line {
					u.statusLine.SetText(line)
				}
				if u.portalBadge != nil && u.portalBadge.Text != badge {
					u.portalBadge.SetText(badge)
				}
			})
		}
	}
}

func (u *ui) refreshLogLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(2 * time.Second):
			tail := readLogTail(filepath.Join(u.cfgDir, "log"), 40)
			if u.log != nil && u.log.Text != tail {
				fyne.Do(func() { u.log.SetText(tail) })
			}
		}
	}
}

func (u *ui) statusInfo() (string, string) {
	ssid, err := u.wifi.ActiveSSID()
	if err != nil || ssid == "" {
		return "Wi-Fi: Disconnected", "Status: Offline"
	}

	online, _ := u.portal.Livecheck(context.Background())
	var portalState string
	if online {
		portalState = "Portal: Authenticated (Online)"
	} else {
		portalState = "Portal: Sign-in Required (Offline)"
	}

	line := fmt.Sprintf("Wi-Fi: Connected to %q", ssid)
	if u.cfg.Paused {
		portalState += " [Auto-login paused]"
	}
	return line, portalState
}

func readLogTail(path string, n int) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	lines := strings.Split(strings.TrimRight(string(b), "\n"), "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return strings.Join(lines, "\n")
}

func (u *ui) creds() (string, string, error) {
	fp, err := config.MachineFingerprint()
	if err != nil {
		return "", "", err
	}
	return u.cfg.GetActiveCreds(fp)
}

func (u *ui) doSignIn() {
	go func() {
		user, pass, err := u.creds()
		if err != nil {
			fyne.Do(func() { u.toast("No credentials stored: " + err.Error()) })
			return
		}
		ok, msg, err := u.portal.Login(context.Background(), user, pass)
		switch {
		case err != nil:
			fyne.Do(func() { u.toast("Sign-in failed: " + err.Error()) })
		case ok:
			fyne.Do(func() {
				u.toast("Signed in successfully!")
				u.refreshStatusOnce()
			})
		default:
			fyne.Do(func() { u.toast("Portal response: " + msg) })
		}
	}()
}

func (u *ui) doSignOut() {
	go func() {
		user, _, err := u.creds()
		if err != nil {
			fyne.Do(func() { u.toast("No credentials stored: " + err.Error()) })
			return
		}
		if err := u.portal.Logout(context.Background(), user); err != nil {
			fyne.Do(func() { u.toast("Sign-out failed: " + err.Error()) })
			return
		}
		fyne.Do(func() {
			u.toast("Signed out")
			u.refreshStatusOnce()
		})
	}()
}
