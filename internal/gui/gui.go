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

	"github.com/achar-pranav/captive-bypass/backends/nmcli"
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
	wifi   *nmcli.Backend
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
	firstRun := err == config.ErrNoConfig || len(cfg.Creds.Ciphertext) == 0

	u := &ui{
		a:      app.NewWithID("io.github.achar-pranav.captive-bypass"),
		cfg:    cfg,
		cfgDir: dir,
		portal: portal.New(cfg.Portal, nil),
		wifi:   nmcli.New(),
	}
	u.w = u.a.NewWindow("captive-bypass")
	u.w.Resize(fyne.NewSize(520, 480))
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

func (u *ui) showMain() {
	u.status = widget.NewLabel("Checking status…")
	u.log = widget.NewMultiLineEntry()
	u.log.Disable()

	signIn := widget.NewButton("Sign in", u.doSignIn)
	signOut := widget.NewButton("Sign out", u.doSignOut)
	pause := widget.NewCheck("Paused — manual portal use", func(on bool) {
		u.cfg.Paused = on
		u.saveConfig()
	})
	pause.SetChecked(u.cfg.Paused)

	ssidsBtn := widget.NewButton("Manage SSIDs", u.showSSIDEditor)
	credsBtn := widget.NewButton("Change credentials", u.showCredsDialog)
	installLabel := widget.NewLabel("Background watcher:")
	watcherStatus := widget.NewLabel("")
	enableBtn := widget.NewButton("Enable", func() {
		go func() {
			err := install.Enable()
			msg := "Watcher enabled — auto sign-in/out active"
			if err != nil {
				msg = "Enable failed: " + err.Error()
			}
			fyne.Do(func() { u.toast(msg); watcherStatus.Refresh() })
		}()
	})
	disableBtn := widget.NewButton("Disable", func() {
		go func() {
			err := install.Disable()
			msg := "Watcher disabled"
			if err != nil {
				msg = "Disable failed: " + err.Error()
			}
			fyne.Do(func() { u.toast(msg); watcherStatus.Refresh() })
		}()
	})
	go func() {
		for {
			on, _ := install.Status()
			txt := "not installed"
			if on {
				txt = "installed"
			}
			fyne.Do(func() { watcherStatus.SetText(txt) })
			time.Sleep(5 * time.Second)
		}
	}()

	top := container.NewVBox(
		u.status,
		container.NewHBox(signIn, signOut, pause),
		widget.NewLabelWithStyle("Registered SSIDs: "+fmt.Sprint(len(u.cfg.SSIDs)), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewHBox(ssidsBtn, credsBtn),
		widget.NewSeparator(),
		container.NewHBox(installLabel, watcherStatus, enableBtn, disableBtn),
		widget.NewSeparator(),
		widget.NewLabel("Activity"),
	)
	u.w.SetContent(container.NewBorder(top, nil, nil, nil, u.log))

	go u.refreshStatusLoop()
	go u.refreshLogLoop()
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
		line += " (paused)"
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
	return u.cfg.GetCreds(fp)
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
