package gui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/achar-pranav/captive-bypass/backends"
)

func newSSIDPicker(wifi backends.Backend, selected map[string]bool, onChange func(string, bool)) fyne.CanvasObject {
	status := widget.NewLabel("Scanning for networks…")
	refresh := widget.NewButton("Refresh", nil)
	list := container.NewVBox()
	load := func() {
		status.SetText("Scanning for networks…")
		refresh.Disable()
		go func() {
			aps, err := wifi.Scan()
			aps = backends.Consolidate(aps)
			fyne.Do(func() {
				refresh.Enable()
				if err != nil {
					status.SetText("Scan failed: " + err.Error())
					list.Objects = nil
					list.Refresh()
					return
				}
				if len(aps) == 0 {
					status.SetText("No networks found")
				} else {
					status.SetText(fmt.Sprintf("%d networks found", len(aps)))
				}
				list.Objects = nil
				for _, ap := range aps {
					ap := ap
					c := widget.NewCheck(ap.SSID, nil)
					c.SetChecked(selected[ap.SSID])
					c.OnChanged = func(on bool) { onChange(ap.SSID, on) }
					list.Add(c)
				}
				list.Refresh()
			})
		}()
	}
	refresh.OnTapped = load
	load()
	return container.NewBorder(container.NewHBox(status, refresh), nil, nil, nil, container.NewScroll(list))
}
