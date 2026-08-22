package gui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/achar-pranav/captive-bypass/internal/config"
)

// staged selection shared by wizard and editor dialogs
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

// ---------- first-run wizard ----------

func (u *ui) showWizard() {
	name := widget.NewEntry()
	name.SetText("default")
	user := widget.NewEntry()
	user.SetPlaceHolder("SRN (portal username)")
	pass := widget.NewPasswordEntry()
	pass.SetPlaceHolder("Password")

	staging := newStaging(u.cfg.SSIDs)
	picker := newSSIDPicker(u.wifi, pickerHooks{
		checked: func(ssid string) bool { return staging.picked[ssid] },
		onRow:   staging.toggle,
	})

	save := widget.NewButton("Save and finish", func() {
		setName := name.Text
		if setName == "" {
			setName = "default"
		}
		if user.Text == "" || pass.Text == "" || len(staging.order) == 0 {
			u.toast("Fill in SRN, password, and tick at least one network")
			return
		}
		fp, err := config.MachineFingerprint()
		if err != nil {
			u.toast("Fingerprint error: " + err.Error())
			return
		}
		if err := u.cfg.SetCredSet(fp, setName, user.Text, pass.Text); err != nil {
			u.toast("Could not encrypt credentials: " + err.Error())
			return
		}
		u.cfg.SSIDs = staging.order
		u.saveConfig()
		u.showMain()
	})

	root := container.NewBorder(nil, save, nil, nil, container.NewVBox(
		widget.NewLabelWithStyle("Welcome — set up captive-bypass", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewForm(
			widget.NewFormItem("Set name", name),
			widget.NewFormItem("Portal username", user),
			widget.NewFormItem("Portal password", pass),
		),
		picker.root,
	))
	u.w.SetContent(root)
}

// ---------- SSID editor (batch apply on OK) ----------

func (u *ui) showSSIDEditor() {
	staging := newStaging(u.cfg.SSIDs)
	picker := newSSIDPicker(u.wifi, pickerHooks{
		checked: func(ssid string) bool { return staging.picked[ssid] },
		onRow:   staging.toggle,
	})
	d := dialog.NewCustomConfirm("Registered SSIDs", "OK", "Cancel", picker.root, func(ok bool) {
		if !ok {
			return
		}
		u.cfg.SSIDs = staging.order
		u.saveConfig()
	}, u.w)
	d.Resize(fyne.NewSize(420, 480))
	d.Show()
}

// ---------- credential set form ----------

func (u *ui) showCredsDialog() {
	name := widget.NewEntry()
	name.SetPlaceHolder("set name")
	user := widget.NewEntry()
	pass := widget.NewPasswordEntry()
	existing := u.activeSet()
	if existing != nil {
		name.SetText(existing.Name)
		user.SetText(existing.Username)
	} else if u.cfg.ActiveSet != "" || len(u.cfg.CredSets) > 0 {
		name.SetText(nextName(u.cfg))
	}
	form := dialog.NewForm("Credential set", "Save", "Cancel", []*widget.FormItem{
		widget.NewFormItem("Set name", name),
		widget.NewFormItem("Username", user),
		widget.NewFormItem("Password", pass),
	}, func(ok bool) {
		if !ok || user.Text == "" || pass.Text == "" {
			return
		}
		fp, err := config.MachineFingerprint()
		if err != nil {
			u.toast("Fingerprint error: " + err.Error())
			return
		}
		nm := name.Text
		if nm == "" {
			nm = nextName(u.cfg)
		}
		if err := u.cfg.SetCredSet(fp, nm, user.Text, pass.Text); err != nil {
			u.toast("Could not encrypt credentials: " + err.Error())
			return
		}
		u.saveConfig()
		u.toast("Credentials saved (" + nm + ")")
		u.showMain()
	}, u.w)
	form.Resize(fyne.NewSize(420, 300))
	form.Show()
}

func (u *ui) activeSet() *config.CredSet {
	for i := range u.cfg.CredSets {
		if u.cfg.CredSets[i].Name == u.cfg.ActiveSet {
			return &u.cfg.CredSets[i]
		}
	}
	return nil
}

func nextName(c *config.Config) string {
	for i := 1; ; i++ {
		cand := fmt.Sprintf("set-%d", i)
		taken := false
		for _, cs := range c.CredSets {
			if cs.Name == cand {
				taken = true
				break
			}
		}
		if !taken {
			return cand
		}
	}
}
