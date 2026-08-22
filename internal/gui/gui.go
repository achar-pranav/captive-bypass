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
	a      fyne.App
	w      fyne.Window
	cfg    *config.Config
	cfgDir string
	portal *portal.Client
	wifi   backends.Backend
	log    *widget.Entry
	status *widget.Label
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

	signIn := widget.NewButton("Sign in", u.doSignIn)
	signOut := widget.NewButton("Sign out", u.doSignOut)

	credsCard := u.cardCredentials()
	netsCard := u.cardNetworks()
	behaviorCard := u.cardBehavior()
	watcherCard := u.cardWatcher()

	body := container.NewVBox(
		container.NewHBox(u.status, signIn, signOut),
		credsCard,
		netsCard,
		behaviorCard,
		watcherCard,
		widget.NewLabel("Activity"),
		u.log,
	)
	u.w.SetContent(container.NewBorder(nil, nil, nil, nil, container.NewScroll(body)))

	go u.refreshStatusLoop()
	go u.refreshLogLoop()
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
	autoLogin := widget.NewCheck("Auto Login (master switch)", func(on bool) {
		u.cfg.Paused = !on
		u.saveConfig()
	})
	autoLogin.SetChecked(!u.cfg.Paused)
	vanguard := widget.NewCheck("Vanguard (experimental)", func(on bool) {
		u.cfg.Vanguard = on
		u.saveConfig()
	})
	vanguard.SetChecked(u.cfg.Vanguard)
	body := container.NewVBox(sectionTitle("Behavior"), autoLogin, vanguard)
	return widget.NewCard("", "", body)
}

func (u *ui) cardWatcher() fyne.CanvasObject {
	watcherStatus := widget.NewLabel("")
	enableBtn := widget.NewButton("Enable", func() {
		go func() {
			err := install.Enable()
			msg := "Watcher enabled — auto sign-in/out active"
			if err != nil {
				msg = "Enable failed: " + err.Error()
			}
			fyne.Do(func() { u.toast(msg) })
		}()
	})
	disableBtn := widget.NewButton("Disable", func() {
		go func() {
			err := install.Disable()
			msg := "Watcher disabled"
			if err != nil {
				msg = "Disable failed: " + err.Error()
			}
			fyne.Do(func() { u.toast(msg) })
		}()
	})
	go func() {
		for i := 0; ; i++ {
			on, _ := install.Status()
			txt := "not installed"
			if on {
				txt = "installed"
			}
			snapshot := txt
			fyne.Do(func() { watcherStatus.SetText(snapshot) })
			time.Sleep(5 * time.Second)
			_ = i
		}
	}()
	body := container.NewVBox(sectionTitle("Background watcher"),
		container.NewHBox(watcherStatus, enableBtn, disableBtn))
	return widget.NewCard("", "", body)
}

func (u *ui) refreshStatusLoop() {
	for {
		txt := u.statusText()
		fyne.Do(func() { u.status.SetText(txt) })
		time.Sleep(3 * time.Second)
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

func (u *ui) refreshLogLoop() {
	for {
		tail := readLogTail(filepath.Join(u.cfgDir, "log"), 40)
		fyne.Do(func() { u.log.SetText(tail) })
		time.Sleep(2 * time.Second)
	}
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
