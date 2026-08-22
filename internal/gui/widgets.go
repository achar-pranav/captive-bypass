package gui

import (
	"image/color"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

// signalBars renders a 4-bar wifi strength glyph; pct is 0-100 (RSSI-scaled).
func signalBars(pct int) fyne.CanvasObject {
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	lit := int(math.Ceil(float64(pct) / 25))
	if lit < 1 {
		lit = 1
	}
	barW, gap, baseY := float32(4), float32(2), float32(20)
	var objects []fyne.CanvasObject
	for i := 0; i < 4; i++ {
		h := float32(5 + i*5)
		c := canvas.NewRectangle(colTextDim)
		c.CornerRadius = 1.5
		if i < lit {
			c.FillColor = colBlue
		} else {
			c.FillColor = color.NRGBA{R: 0x24, G: 0x28, B: 0x2D, A: 0xFF}
		}
		x := float32(i) * (barW + gap)
		c.Move(fyne.NewPos(x, baseY-h))
		c.Resize(fyne.NewSize(barW, h))
		objects = append(objects, c)
	}
	box := container.NewWithoutLayout(objects...)
	box.Resize(fyne.NewSize(4*barW+3*gap, baseY))
	return container.NewPadded(box)
}
