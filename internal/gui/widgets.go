package gui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// signalBadge renders a clean, standard text glyph badge representing Wi-Fi strength.
func signalBadge(pct int) fyne.CanvasObject {
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	bars := "▂▄▆█"
	switch {
	case pct < 25:
		bars = "▂   "
	case pct < 50:
		bars = "▂▄  "
	case pct < 75:
		bars = "▂▄▆ "
	}
	lbl := widget.NewLabel(fmt.Sprintf("%s %d%%", bars, pct))
	lbl.TextStyle = fyne.TextStyle{Monospace: true}
	return lbl
}
