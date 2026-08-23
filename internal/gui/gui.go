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
	a            fyne.App
	w            fyne.Window
	cfg          *config.Config
	cfgDir       string
	portal       *portal.Client
	wifi         backends.Backend
	log          *widget.Entry
	status       *widget.Label
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
	u.w.Resize(fyne.NewSize(380, 620))
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
	u.status = widget.NewLabel("Checking status…")
	u.log = widget.NewMultiLineEntry()
	u.log.Disable()
	u.log.Wrapping = fyne.TextWrapBreak
	u.log.TextStyle = fyne.TextStyle{Monospace: true}

	credsCard := u.cardCredentials()
	netsCard := u.cardNetworks()
	behaviorCard := u.cardBehavior()
	watcherCard := u.cardWatcher()

	activityBox := container.NewVScroll(u.log)
	activityBox.SetMinSize(fyne.NewSize(0, 160))
	activityCard := widget.NewCard("", "Activity", activityBox)

	body := container.NewVBox(
		u.status,
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
	go u.refreshStatusLoop(ctx)
	go u.refreshLogLoop(ctx)
}

func sectionTitle(s string) fyne.CanvasObject {
	l := widget.NewLabel(s)
	l.TextStyle = fyne.TextStyle{Bold: true}
	return l
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

	addBtn := widget.NewButton("Add / update", u.showCredsDialog)
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
	hint := widget.NewLabel("The highlighted set is used for sign-in.")
	hint.TextStyle = fyne.TextStyle{Italic: true}

	body := container.NewVBox(sectionTitle("Credential sets"), radio, hint,
		container.NewHBox(addBtn, delBtn))
	if len(names) == 0 {
		body = container.NewVBox(sectionTitle("Credential sets"),
			widget.NewLabel("None yet — add one so sign-in has something to use."),
			addBtn)
	}
	return widget.NewCard("", "", body)
}

func (u *ui) cardNetworks() fyne.CanvasObject {
	list := container.NewVBox()
	for _, s := range u.cfg.SSIDs {
		list.Add(container.NewHBox(widget.NewLabel("•  " + s)))
	}
	if len(u.cfg.SSIDs) == 0 {
		list.Add(widget.NewLabel("None registered yet."))
	}
	manageBtn := widget.NewButton("Manage SSIDs", func() { u.showSSIDEditor() })
	body := container.NewVBox(sectionTitle("Recognized networks"), list, manageBtn)
	return widget.NewCard("", "", body)
}

func (u *ui) cardBehavior() fyne.CanvasObject {
	autoLoginBtn := widget.NewButton("", nil)
	autoLoginBtn.OnTapped = func() {
		u.cfg.Paused = !u.cfg.Paused
		u.saveConfig()
		autoLoginBtn.SetText(map[bool]string{true: "Auto Login: ON", false: "Auto Login: OFF"}[!u.cfg.Paused])
	}
	autoLoginBtn.SetText(map[bool]string{true: "Auto Login: ON", false: "Auto Login: OFF"}[!u.cfg.Paused])

	vanguardBtn := widget.NewButton("", nil)
	vanguardBtn.OnTapped = func() {
		u.cfg.Vanguard = !u.cfg.Vanguard
		u.saveConfig()
		vanguardBtn.SetText(map[bool]string{true: "Vanguard: ON", false: "Vanguard: OFF"}[u.cfg.Vanguard])
	}
	vanguardBtn.SetText(map[bool]string{true: "Vanguard: ON", false: "Vanguard: OFF"}[u.cfg.Vanguard])

	body := container.NewVBox(sectionTitle("Behavior"), autoLoginBtn, vanguardBtn)
	return widget.NewCard("", "", body)
}

func (u *ui) cardWatcher() fyne.CanvasObject {
	watcherStatus := widget.NewLabel("checking…")
	enableBtn := widget.NewButton("Enable", func() {
		go func() {
			err := install.Enable()
			msg := "Auto-start enabled — sign-in/out runs in background"
			if err != nil {
				msg = "Enable failed: " + err.Error()
			}
			fyne.Do(func() { u.toast(msg) })
		}()
	})
	disableBtn := widget.NewButton("Disable", func() {
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
			txt := "off"
			if on {
				txt = "on"
			}
			if watcherStatus.Text != txt {
				fyne.Do(func() { watcherStatus.SetText(txt) })
			}
			time.Sleep(5 * time.Second)
		}
	}()
	body := container.NewVBox(
		sectionTitle("Run automatically at login"),
		widget.NewLabel("Installs a small background service so sign-in/out happens without opening this app."),
		container.NewHBox(watcherStatus, enableBtn, disableBtn),
	)
	return widget.NewCard("", "", body)
}

func (u *ui) refreshStatusLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(3 * time.Second):
			txt := u.statusText()
			if u.status.Text != txt {
				fyne.Do(func() { u.status.SetText(txt) })
			}
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
			if u.log.Text != tail {
				fyne.Do(func() { u.log.SetText(tail) })
			}
		}
	}
}

func (u *ui) statusText() string {
	ssid, err := u.wifi.ActiveSSID()
	if err != nil || ssid == "" {
		return "Not connected to any WiFi"
	}
	online, _ := u.portal.Livecheck(context.Background())
	state := "offline"
	if online {
		state = "online"
	}
	line := fmt.Sprintf("WiFi: %s — portal %s", ssid, state)
	if u.cfg.Paused {
		line += " (auto-login off)"
	}
	return line
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
			fyne.Do(func() { u.toast("Signed in") })
		default:
			fyne.Do(func() { u.toast("Portal says: " + msg) })
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
		fyne.Do(func() { u.toast("Signed out") })
	}()
}
