package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/achar-pranav/captive-bypass/backends"
)

type pickerHooks struct {
	checked func(ssid string) bool
	onRow   func(ssid string, on bool)
}

type ssidPicker struct {
	root    fyne.CanvasObject
	status  *widget.Label
	refresh *widget.Button
	list    *fyne.Container
	wifi    backends.Backend
	hooks   pickerHooks
	spam    antiSpam
}

func newSSIDPicker(wifi backends.Backend, hooks pickerHooks) *ssidPicker {
	p := &ssidPicker{wifi: wifi, hooks: hooks}
	p.status = widget.NewLabel("Networks")
	p.status.TextStyle = fyne.TextStyle{Bold: true}

	// Small refresh button next to title: "Networks ⟳"
	p.refresh = widget.NewButton("Networks ⟳", func() {
		p.spam.Run(p.load)
	})

	p.list = container.NewVBox()

	titleRow := container.NewHBox(
		p.status,
		layout.NewSpacer(),
		p.refresh,
	)

	p.root = container.NewBorder(
		titleRow,
		nil, nil, nil,
		container.NewScroll(p.list),
	)

	p.load()
	return p
}

func (p *ssidPicker) load() {
	p.refresh.Disable()
	go func() {
		aps, err := p.wifi.Scan()
		aps = backends.Consolidate(aps)
		fyne.Do(func() {
			p.refresh.Enable()
			p.renderRows(aps, err)
		})
	}()
}

func (p *ssidPicker) renderRows(aps []backends.AP, err error) {
	p.list.Objects = nil
	if err != nil {
		p.list.Add(widget.NewLabel("Scan error: " + err.Error()))
		p.list.Refresh()
		return
	}
	if len(aps) == 0 {
		p.list.Add(widget.NewLabel("No nearby Wi-Fi networks found"))
		p.list.Refresh()
		return
	}

	for _, ap := range aps {
		ap := ap
		c := widget.NewCheck(ap.SSID, nil)
		c.SetChecked(p.hooks.checked(ap.SSID))
		c.OnChanged = func(on bool) { p.hooks.onRow(ap.SSID, on) }

		row := container.NewBorder(
			nil, nil,
			c,
			signalLabel(ap.Signal),
		)
		p.list.Add(row)
	}
	p.list.Refresh()
}
