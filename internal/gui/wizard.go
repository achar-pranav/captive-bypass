package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/achar-pranav/captive-bypass/internal/config"
)

type ssidStaging struct {
	picked map[string]bool
	order  []string
}

func newStaging(ssids []string) *ssidStaging {
	s := &ssidStaging{picked: map[string]bool{}}
	for _, v := range ssids {
		if !s.picked[v] {
			s.picked[v] = true
			s.order = append(s.order, v)
		}
	}
	return s
}

func (s *ssidStaging) toggle(ssid string, on bool) {
	if on {
		if !s.picked[ssid] {
			s.picked[ssid] = true
			s.order = append(s.order, ssid)
		}
		return
	}
	if s.picked[ssid] {
		delete(s.picked, ssid)
		for i, v := range s.order {
			if v == ssid {
				s.order = append(s.order[:i], s.order[i+1:]...)
				break
			}
		}
	}
}

// ---------- Screen 1: Permissions & Trust ----------

func (u *ui) showWizardScreen1() {
	title := widget.NewLabel("Permissions & Trust")
	title.TextStyle = fyne.TextStyle{Bold: true}

	info := widget.NewLabel(
		"Welcome to captive-bypass.\n\n" +
			"This tool automates logging into the campus captive portal whenever you connect.\n\n" +
			"• Hardware & Wi-Fi: Reads the active network SSID without requiring root privileges.\n" +
			"• Open Source: 100% auditable code (github.com/achar-pranav/captive-bypass).\n" +
			"• Local Only: Your portal credentials never leave your machine.\n\n" +
			"Click Continue to set up credentials, or Skip to jump to the main menu.",
	)
	info.Wrapping = fyne.TextWrapWord

	continueBtn := widget.NewButton("Continue", func() {
		u.spam.Run(func() {
			fyne.Do(u.showWizardScreen2)
		})
	})
	continueBtn.Importance = widget.HighImportance

	skipBtn := widget.NewButton("Skip", func() {
		u.spam.Run(func() {
			fyne.Do(u.showMain)
		})
	})

	buttons := container.NewVBox(
		continueBtn,
		skipBtn,
	)

	body := container.NewVBox(
		title,
		layout.NewSpacer(),
		info,
		layout.NewSpacer(),
		buttons,
	)

	u.w.SetContent(u.wrapWithToast(body))
}

// ---------- Screen 2: Add one set of credentials ----------

func (u *ui) showWizardScreen2() {
	title := widget.NewLabel("Add Credentials")
	title.TextStyle = fyne.TextStyle{Bold: true}

	nameEntry := widget.NewEntry()
	nameEntry.SetText("default")

	userEntry := widget.NewEntry()
	userEntry.SetPlaceHolder("PES1UG...")

	passEntry, passRow := newPasswordEntryWithToggle("Portal Password")

	disclaimer := widget.NewLabel(
		"Passwords are never stored in plaintext. We use OS hardware fingerprinting " +
			"and AES-GCM encryption to prevent theft by copy.",
	)
	disclaimer.TextStyle = fyne.TextStyle{Italic: true}
	disclaimer.Wrapping = fyne.TextWrapWord

	backBtn := widget.NewButton("Back", func() {
		u.spam.Run(func() {
			fyne.Do(u.showWizardScreen1)
		})
	})

	// Continue button takes the entire width of the bottom
	continueBtn := widget.NewButton("Continue", func() {
		u.spam.Run(func() {
			user := userEntry.Text
			pass := passEntry.Text
			name := nameEntry.Text
			if name == "" {
				name = "default"
			}
			if user == "" || pass == "" {
				u.toast("Please fill in your SRN and Password")
				return
			}
			fp, err := config.MachineFingerprint()
			if err != nil {
				u.toast("Fingerprint error: " + err.Error())
				return
			}
			if err := u.cfg.SetCredSet(fp, name, user, pass); err != nil {
				u.toast("Could not encrypt credentials: " + err.Error())
				return
			}
			u.saveConfig()
			fyne.Do(u.showWizardScreen3)
		})
	})
	continueBtn.Importance = widget.HighImportance

	form := container.NewVBox(
		widget.NewLabel("1. Credentials name"),
		nameEntry,
		widget.NewLabel("2. Username"),
		userEntry,
		widget.NewLabel("3. Password"),
		passRow,
		disclaimer,
	)

	btns := container.NewVBox(
		backBtn,
		continueBtn,
	)

	body := container.NewBorder(
		title,
		btns,
		nil, nil,
		form,
	)

	u.w.SetContent(u.wrapWithToast(body))
}

// ---------- Screen 3: Select at least one SSID to log into automatically ----------

func (u *ui) showWizardScreen3() {
	title := widget.NewLabel("Select at least one SSID to log into automatically")
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Wrapping = fyne.TextWrapWord

	staging := newStaging(u.cfg.SSIDs)
	picker := newSSIDPicker(u.wifi, pickerHooks{
		checked: func(ssid string) bool { return staging.picked[ssid] },
		onRow:   staging.toggle,
	})

	backBtn := widget.NewButton("Back", func() {
		u.spam.Run(func() {
			fyne.Do(u.showWizardScreen2)
		})
	})

	// Done button takes the entire width of the bottom
	doneBtn := widget.NewButton("Done", func() {
		u.spam.Run(func() {
			if len(staging.order) == 0 {
				u.toast("Please select at least one network")
				return
			}
			u.cfg.SSIDs = staging.order
			u.saveConfig()
			fyne.Do(u.showMain)
		})
	})
	doneBtn.Importance = widget.HighImportance

	btns := container.NewVBox(
		backBtn,
		doneBtn,
	)

	body := container.NewBorder(
		title,
		btns,
		nil, nil,
		picker.root,
	)

	u.w.SetContent(u.wrapWithToast(body))
}
