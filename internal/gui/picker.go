package gui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/achar-pranav/captive-bypass/backends"
)

// pickerHooks let wizard and editor share the scan/refresh UI while
// deciding their own checked-state and toggle behavior.
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
}

func newSSIDPicker(wifi backends.Backend, hooks pickerHooks) *ssidPicker {
	p := &ssidPicker{wifi: wifi, hooks: hooks}
	p.status = widget.NewLabel("Scanning for networks…")
	p.refresh = widget.NewButton("Refresh", nil)
	p.list = container.NewVBox()
	p.refresh.OnTapped = p.load
	p.root = container.NewBorder(
		container.NewHBox(p.status, p.refresh),
		nil, nil, nil,
		container.NewScroll(p.list),
	)
	p.load()
	return p
}

func (p *ssidPicker) load() {
	p.status.SetText("Scanning for networks…")
	p.refresh.Disable()
	go func() {
		aps, err := p.wifi.Scan()
		aps = backends.Consolidate(aps)
		fyne.Do(func() {
			p.refresh.Enable()
			switch {
			case err != nil:
				p.status.SetText("Scan failed: " + err.Error())
				p.list.Objects = nil
			case len(aps) == 0:
				p.status.SetText("No networks found")
			default:
				p.status.SetText(fmt.Sprintf("%d networks found", len(aps)))
			}
			p.renderRows(aps)
		})
	}()
}

func (p *ssidPicker) renderRows(aps []backends.AP) {
	p.list.Objects = nil
	for _, ap := range aps {
		ap := ap
		c := widget.NewCheck("", nil)
		c.SetChecked(p.hooks.checked(ap.SSID))
		c.OnChanged = func(on bool) { p.hooks.onRow(ap.SSID, on) }
		label := widget.NewLabel(ap.SSID)
		label.Truncation = fyne.TextTruncateEllipsis
		row := container.NewBorder(nil, nil,
			container.NewHBox(signalBars(ap.Signal), label),
			c,
		)
		p.list.Add(row)
	}
	p.list.Refresh()
}
