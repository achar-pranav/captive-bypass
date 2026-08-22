package gui

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/achar-pranav/captive-bypass/internal/config"
)

func cleanSSID(s string) string {
	return strings.TrimSpace(s)
}

func (u *ui) showWizard() {
	user := widget.NewEntry()
	user.SetPlaceHolder("SRN (portal username)")
	pass := widget.NewPasswordEntry()
	pass.SetPlaceHolder("Password")
	ssid := widget.NewEntry()
	ssid.SetPlaceHolder("e.g. ELEMENT BLOCK")

	list := container.NewVBox()
	var render func()
	render = func() {
		list.Objects = nil
		for _, s := range u.cfg.SSIDs {
			name := s
			list.Add(container.NewHBox(widget.NewLabel(name), widget.NewButton("Remove", func() {
				u.cfg.SSIDs = removeSSID(u.cfg.SSIDs, name)
				render()
			})))
		}
		list.Refresh()
	}
	addBtn := widget.NewButton("Add SSID", func() {
		name := cleanSSID(ssid.Text)
		if name == "" {
			u.toast("Type a WiFi network name first")
			return
		}
		for _, s := range u.cfg.SSIDs {
			if s == name {
				ssid.SetText("")
				return
			}
		}
		u.cfg.SSIDs = append(u.cfg.SSIDs, name)
		ssid.SetText("")
		render()
	})
	render()

	save := widget.NewButton("Save and finish", func() {
		if user.Text == "" || pass.Text == "" || len(u.cfg.SSIDs) == 0 {
			u.toast("Fill in SRN, password, and at least one SSID")
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

	root := container.NewVBox(
		widget.NewLabelWithStyle("Welcome — set up captive-bypass", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewForm(
			widget.NewFormItem("Portal username", user),
			widget.NewFormItem("Portal password", pass),
		),
		container.NewHBox(ssid, addBtn),
		list,
		save,
	)
	u.w.SetContent(root)
	baseRender := render
	render = func() {
		baseRender()
		root.Refresh()
	}
}

func (u *ui) showSSIDEditor() {
	entry := widget.NewEntry()
	entry.SetPlaceHolder("SSID to add")
	list := container.NewVBox()
	var render func()
	render = func() {
		list.Objects = nil
		for _, s := range u.cfg.SSIDs {
			name := s
			list.Add(container.NewHBox(widget.NewLabel(name), widget.NewButton("Remove", func() {
				u.cfg.SSIDs = removeSSID(u.cfg.SSIDs, name)
				u.saveConfig()
				render()
			})))
		}
		list.Refresh()
	}
	doAdd := func() {
		name := cleanSSID(entry.Text)
		if name == "" {
			u.toast("Type a WiFi network name first")
			return
		}
		for _, s := range u.cfg.SSIDs {
			if s == name {
				entry.SetText("")
				return
			}
		}
		u.cfg.SSIDs = append(u.cfg.SSIDs, name)
		u.saveConfig()
		entry.SetText("")
		render()
	}
	addBtn := widget.NewButton("Add", doAdd)
	entry.OnSubmitted = func(string) { doAdd() }
	render()
	content := container.NewVBox(
		entry,
		addBtn,
		container.NewScroll(list),
	)
	d := dialog.NewCustom("Registered SSIDs", "Done", content, u.w)
	d.Resize(fyne.NewSize(400, 360))
	d.Show()
}

func (u *ui) showCredsDialog() {
	user := widget.NewEntry()
	pass := widget.NewPasswordEntry()
	existingUser := u.cfg.Creds.Username
	user.SetText(existingUser)
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
