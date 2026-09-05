package gui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/achar-pranav/captive-bypass/backends"
	"github.com/achar-pranav/captive-bypass/backends/auto"
	"github.com/achar-pranav/captive-bypass/internal/config"
	"github.com/achar-pranav/captive-bypass/internal/portal"
)

type ui struct {
	a             fyne.App
	w             fyne.Window
	cfg           *config.Config
	cfgDir        string
	portal        *portal.Client
	wifi          backends.Backend
	led           *ledWidget
	statusText    *widget.Label
	toastWidget   *bottomToast
	refreshCancel context.CancelFunc
	spam          antiSpam
	mu            sync.Mutex
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
		a:           app.NewWithID("io.github.achar-pranav.captive-bypass"),
		cfg:         cfg,
		cfgDir:      dir,
		portal:      portal.New(cfg.Portal, nil),
		wifi:        auto.Default(),
		led:         newLEDWidget(LEDRed),
		statusText:  widget.NewLabel("<Disconnected>"),
		toastWidget: newBottomToast(),
	}

	u.a.Settings().SetTheme(newAmoledTheme())
	u.w = u.a.NewWindow("captive-bypass")
	u.w.Resize(fyne.NewSize(420, 560))

	if firstRun {
		u.showWizardScreen1()
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

// toast displays a bottom non-intrusive popup instead of a center dialog
func (u *ui) toast(msg string) {
	if u.toastWidget != nil {
		u.toastWidget.Show(msg)
	}
}

func (u *ui) wrapWithToast(content fyne.CanvasObject) fyne.CanvasObject {
	return container.NewStack(
		content,
		u.toastWidget.container,
	)
}

// ---------- Main Menu ----------

func (u *ui) showMain() {
	u.statusText.TextStyle = fyne.TextStyle{Bold: true}

	// 1. Top status line with 4-state LED and title text <Status>
	statusHeader := container.NewHBox(
		u.led,
		u.statusText,
	)

	// 2. Title card 1: Networks
	networksCard := u.cardNetworks()

	// 3. Title card 2: Creds
	credsCard := u.cardCreds()

	// 4. Bottom two checkboxes
	toggleBypass := widget.NewCheck("Enable/Disable the captive-bypass", func(on bool) {
		u.cfg.Paused = !on
		u.saveConfig()
		if !on {
			u.setLED(LEDRed, "<Disabled>")
		} else {
			u.refreshStatusOnce()
		}
	})
	toggleBypass.SetChecked(!u.cfg.Paused)

	toggleVanguard := widget.NewCheck("Enable/Disable Vanguard(experimental)", func(on bool) {
		u.cfg.Vanguard = on
		u.saveConfig()
		if on {
			u.toast("Vanguard telemetry enabled")
		} else {
			u.toast("Vanguard telemetry disabled")
		}
	})
	toggleVanguard.SetChecked(u.cfg.Vanguard)

	bottomBox := container.NewVBox(
		widget.NewSeparator(),
		toggleBypass,
		toggleVanguard,
	)

	body := container.NewVBox(
		statusHeader,
		layout.NewSpacer(),
		networksCard,
		credsCard,
		layout.NewSpacer(),
		bottomBox,
	)

	u.w.SetContent(u.wrapWithToast(body))

	if u.refreshCancel != nil {
		u.refreshCancel()
	}
	ctx, cancel := context.WithCancel(context.Background())
	u.refreshCancel = cancel

	u.refreshStatusOnce()
	go u.refreshStatusLoop(ctx)
}

func (u *ui) cardNetworks() fyne.CanvasObject {
	title := widget.NewLabel("Networks")
	title.TextStyle = fyne.TextStyle{Bold: true}

	addBtn := widget.NewButton("Add", func() {
		u.spam.Run(func() {
			fyne.Do(u.showAddNetworksDialog)
		})
	})
	addBtn.Importance = widget.HighImportance

	manageBtn := widget.NewButton("Manage", func() {
		u.spam.Run(func() {
			fyne.Do(u.showManageNetworksDialog)
		})
	})

	actions := container.NewHBox(addBtn, manageBtn)

	header := container.NewBorder(
		nil, nil,
		title,
		actions,
	)

	summary := widget.NewLabel(fmt.Sprintf("%d registered", len(u.cfg.SSIDs)))
	summary.TextStyle = fyne.TextStyle{Italic: true}

	content := container.NewVBox(
		header,
		summary,
	)
	return widget.NewCard("", "", content)
}

func (u *ui) cardCreds() fyne.CanvasObject {
	title := widget.NewLabel("Creds")
	title.TextStyle = fyne.TextStyle{Bold: true}

	addBtn := widget.NewButton("Add", func() {
		u.spam.Run(func() {
			fyne.Do(u.showAddCredsDialog)
		})
	})
	addBtn.Importance = widget.HighImportance

	manageBtn := widget.NewButton("Manage", func() {
		u.spam.Run(func() {
			fyne.Do(u.showManageCredsDialog)
		})
	})

	actions := container.NewHBox(addBtn, manageBtn)

	header := container.NewBorder(
		nil, nil,
		title,
		actions,
	)

	activeName := u.cfg.ActiveSet
	if activeName == "" {
		activeName = "None"
	}
	summary := widget.NewLabel(fmt.Sprintf("Active: %s (%d configured)", activeName, len(u.cfg.CredSets)))
	summary.TextStyle = fyne.TextStyle{Italic: true}

	content := container.NewVBox(
		header,
		summary,
	)
	return widget.NewCard("", "", content)
}

// ---------- Networks: Add & Manage Dialogs ----------

func (u *ui) showAddNetworksDialog() {
	staging := newStaging(u.cfg.SSIDs)
	picker := newSSIDPicker(u.wifi, pickerHooks{
		checked: func(ssid string) bool { return staging.picked[ssid] },
		onRow:   staging.toggle,
	})

	var win fyne.Window
	win = u.a.NewWindow("Add Networks")
	win.Resize(fyne.NewSize(400, 480))

	saveBtn := widget.NewButton("Save", func() {
		u.spam.Run(func() {
			u.cfg.SSIDs = staging.order
			u.saveConfig()
			fyne.Do(func() {
				win.Close()
				u.showMain()
				u.toast("Networks updated")
			})
		})
	})
	saveBtn.Importance = widget.HighImportance

	cancelBtn := widget.NewButton("Cancel", func() {
		win.Close()
	})

	btns := container.NewHBox(layout.NewSpacer(), cancelBtn, saveBtn)
	content := container.NewBorder(nil, btns, nil, nil, picker.root)

	win.SetContent(content)
	win.Show()
}

func (u *ui) showManageNetworksDialog() {
	var win fyne.Window
	win = u.a.NewWindow("Manage Networks")
	win.Resize(fyne.NewSize(380, 420))

	rows := container.NewVBox()

	var renderList func()
	renderList = func() {
		rows.Objects = nil
		for _, s := range u.cfg.SSIDs {
			ssid := s
			lbl := widget.NewLabel(ssid)
			lbl.TextStyle = fyne.TextStyle{Monospace: true}

			// Small red square trash can button
			trashBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
				u.spam.Run(func() {
					u.deleteSSID(ssid)
					fyne.Do(func() {
						renderList()
						u.toast("Deleted " + ssid)
					})
				})
			})
			trashBtn.Importance = widget.DangerImportance

			row := container.NewBorder(nil, nil, lbl, trashBtn)
			rows.Add(row)
		}
		if len(u.cfg.SSIDs) == 0 {
			rows.Add(widget.NewLabel("No registered networks."))
		}
		rows.Refresh()
	}

	renderList()

	doneBtn := widget.NewButton("Done", func() {
		win.Close()
		u.showMain()
	})
	doneBtn.Importance = widget.HighImportance

	content := container.NewBorder(
		widget.NewLabel("Delete Registered Networks:"),
		doneBtn,
		nil, nil,
		container.NewScroll(rows),
	)

	win.SetContent(content)
	win.Show()
}

func (u *ui) deleteSSID(ssid string) {
	var remaining []string
	for _, s := range u.cfg.SSIDs {
		if s != ssid {
			remaining = append(remaining, s)
		}
	}
	u.cfg.SSIDs = remaining
	u.saveConfig()
}

// ---------- Creds: Add & Manage Dialogs ----------

func (u *ui) showAddCredsDialog() {
	var win fyne.Window
	win = u.a.NewWindow("Add Credentials")
	win.Resize(fyne.NewSize(380, 360))

	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Profile name (e.g. personal, lab)")

	userEntry := widget.NewEntry()
	userEntry.SetPlaceHolder("SRN")

	passEntry := widget.NewPasswordEntry()
	passEntry.SetPlaceHolder("Password")

	disclaimer := widget.NewLabel(
		"Passwords are never stored in plaintext. We use OS hardware fingerprinting " +
			"and AES-GCM encryption to prevent theft by copy.",
	)
	disclaimer.TextStyle = fyne.TextStyle{Italic: true}
	disclaimer.Wrapping = fyne.TextWrapWord

	saveBtn := widget.NewButton("Save", func() {
		u.spam.Run(func() {
			if userEntry.Text == "" || passEntry.Text == "" {
				fyne.Do(func() { u.toast("Please fill in Username and Password") })
				return
			}
			fp, err := config.MachineFingerprint()
			if err != nil {
				fyne.Do(func() { u.toast("Fingerprint error: " + err.Error()) })
				return
			}
			nm := nameEntry.Text
			if nm == "" {
				nm = fmt.Sprintf("set-%d", len(u.cfg.CredSets)+1)
			}
			if err := u.cfg.SetCredSet(fp, nm, userEntry.Text, passEntry.Text); err != nil {
				fyne.Do(func() { u.toast("Encrypt error: " + err.Error()) })
				return
			}
			u.saveConfig()
			fyne.Do(func() {
				win.Close()
				u.showMain()
				u.toast("Credentials saved: " + nm)
			})
		})
	})
	saveBtn.Importance = widget.HighImportance

	cancelBtn := widget.NewButton("Cancel", func() {
		win.Close()
	})

	form := container.NewVBox(
		widget.NewLabel("1. Credentials name"),
		nameEntry,
		widget.NewLabel("2. Username (SRN)"),
		userEntry,
		widget.NewLabel("3. Password"),
		passEntry,
		disclaimer,
	)

	btns := container.NewHBox(layout.NewSpacer(), cancelBtn, saveBtn)
	content := container.NewBorder(nil, btns, nil, nil, form)

	win.SetContent(content)
	win.Show()
}

func (u *ui) showManageCredsDialog() {
	var win fyne.Window
	win = u.a.NewWindow("Manage Credentials")
	win.Resize(fyne.NewSize(380, 420))

	rows := container.NewVBox()

	var renderList func()
	renderList = func() {
		rows.Objects = nil

		var names []string
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
			u.toast("Active profile: " + chosen)
		})
		radio.SetSelected(u.cfg.ActiveSet)

		for _, cs := range u.cfg.CredSets {
			name := cs.Name
			trashBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
				u.spam.Run(func() {
					if err := u.cfg.DeleteCredSet(name); err != nil {
						fyne.Do(func() { u.toast(err.Error()) })
						return
					}
					u.saveConfig()
					fyne.Do(func() {
						renderList()
						u.toast("Deleted " + name)
					})
				})
			})
			trashBtn.Importance = widget.DangerImportance

			row := container.NewHBox(
				widget.NewLabel(fmt.Sprintf("%s (%s)", name, cs.Username)),
				layout.NewSpacer(),
				trashBtn,
			)
			rows.Add(row)
		}

		if len(u.cfg.CredSets) == 0 {
			rows.Add(widget.NewLabel("No credentials stored."))
		} else {
			rows.Add(widget.NewSeparator())
			rows.Add(widget.NewLabel("Active profile:"))
			rows.Add(radio)
		}
		rows.Refresh()
	}

	renderList()

	doneBtn := widget.NewButton("Done", func() {
		win.Close()
		u.showMain()
	})
	doneBtn.Importance = widget.HighImportance

	content := container.NewBorder(
		widget.NewLabel("Credential Profiles:"),
		doneBtn,
		nil, nil,
		container.NewScroll(rows),
	)

	win.SetContent(content)
	win.Show()
}

// ---------- Status polling & LED Updates ----------

func (u *ui) setLED(state LEDState, text string) {
	fyne.Do(func() {
		if u.led != nil {
			u.led.SetState(state)
		}
		if u.statusText != nil {
			u.statusText.SetText(text)
		}
	})
}

func (u *ui) refreshStatusOnce() {
	go func() {
		state, text := u.evalStatus()
		u.setLED(state, text)
	}()
}

func (u *ui) refreshStatusLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(3 * time.Second):
			state, text := u.evalStatus()
			u.setLED(state, text)
		}
	}
}

func (u *ui) evalStatus() (LEDState, string) {
	if u.cfg.Paused {
		return LEDRed, "<Disabled>"
	}

	ssid, err := u.wifi.ActiveSSID()
	if err != nil || ssid == "" {
		return LEDRed, "<Disconnected>"
	}

	// Check if signal is at edge of network
	aps, scanErr := u.wifi.Scan()
	if scanErr == nil {
		for _, ap := range aps {
			if ap.SSID == ssid {
				threshold := u.cfg.SignalThreshold()
				if ap.Signal > 0 && ap.Signal <= threshold {
					if u.cfg.Vanguard {
						u.toast(fmt.Sprintf("[Vanguard] Weak Wi-Fi (%d%% <= %d%%)", ap.Signal, threshold))
					}
					return LEDOrange, fmt.Sprintf("<Edge of Network: %s (%d%%)>", ssid, ap.Signal)
				}
				break
			}
		}
	}

	online, _ := u.portal.Livecheck(context.Background())
	if online {
		return LEDGreen, fmt.Sprintf("<Connected: %s>", ssid)
	}

	return LEDYellow, fmt.Sprintf("<In Progress: %s>", ssid)
}
