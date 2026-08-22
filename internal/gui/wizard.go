package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/achar-pranav/captive-bypass/internal/config"
)

func (u *ui) toggleSSID(name string, on bool) {
	if on {
		for _, s := range u.cfg.SSIDs {
			if s == name {
				return
			}
		}
		u.cfg.SSIDs = append(u.cfg.SSIDs, name)
		return
	}
	u.cfg.SSIDs = removeSSID(u.cfg.SSIDs, name)
}

func (u *ui) showWizard() {
	user := widget.NewEntry()
	user.SetPlaceHolder("SRN (portal username)")
	pass := widget.NewPasswordEntry()
	pass.SetPlaceHolder("Password")

	picker := newSSIDPicker(u.wifi, ssidSet(u.cfg.SSIDs), u.toggleSSID)

	save := widget.NewButton("Save and finish", func() {
		if user.Text == "" || pass.Text == "" || len(u.cfg.SSIDs) == 0 {
			u.toast("Fill in SRN, password, and tick at least one network")
			return
		}
		fp, err := config.MachineFingerprint()
		if err != nil {
			u.toast("Fingerprint error: " + err.Error())
			return
		}
		if err := u.cfg.SetCreds(fp, user.Text, pass.Text); err != nil {
			u.toast("Could not encrypt credentials: " + err.Error())
			return
		}
		u.saveConfig()
		u.showMain()
	})

	root := container.NewBorder(nil, save, nil, nil, container.NewVBox(
		widget.NewLabelWithStyle("Welcome — set up captive-bypass", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewForm(
			widget.NewFormItem("Portal username", user),
			widget.NewFormItem("Portal password", pass),
		),
		picker,
	))
	u.w.SetContent(root)
}

func (u *ui) showSSIDEditor() {
	picker := newSSIDPicker(u.wifi, ssidSet(u.cfg.SSIDs), func(ssid string, on bool) {
		u.toggleSSID(ssid, on)
		u.saveConfig()
	})
	d := dialog.NewCustom("Registered SSIDs", "Done", picker, u.w)
	d.Resize(fyne.NewSize(420, 420))
	d.Show()
}

func ssidSet(ssids []string) map[string]bool {
	m := make(map[string]bool, len(ssids))
	for _, s := range ssids {
		m[s] = true
	}
	return m
}

func (u *ui) showCredsDialog() {
	user := widget.NewEntry()
	pass := widget.NewPasswordEntry()
	user.SetText(u.cfg.Creds.Username)
	form := dialog.NewForm("Change credentials", "Save", "Cancel", []*widget.FormItem{
		widget.NewFormItem("Username", user),
		widget.NewFormItem("New password", pass),
	}, func(ok bool) {
		if !ok || pass.Text == "" {
			return
		}
		fp, err := config.MachineFingerprint()
		if err != nil {
			u.toast("Fingerprint error: " + err.Error())
			return
		}
		if err := u.cfg.SetCreds(fp, user.Text, pass.Text); err != nil {
			u.toast("Could not encrypt credentials: " + err.Error())
			return
		}
		u.saveConfig()
		u.toast("Credentials updated")
	}, u.w)
	form.Resize(fyne.NewSize(420, 240))
	form.Show()
}

func removeSSID(list []string, name string) []string {
	out := list[:0]
	for _, s := range list {
		if s != name {
			out = append(out, s)
		}
	}
	return out
}
