package gui

import (
	"fmt"
	"image/color"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// LEDState represents the 4 states of the top-left status LED:
// 1. Green - connected
// 2. Red - disconnected
// 3. Yellow - in progress
// 4. Orange - edge of network (signal <= threshold)
type LEDState int

const (
	LEDGreen LEDState = iota
	LEDRed
	LEDYellow
	LEDOrange
)

var (
	colLEDGreen  = color.NRGBA{R: 0x23, G: 0xA5, B: 0x5A, A: 0xFF} // #23A55A
	colLEDRed    = color.NRGBA{R: 0xF2, G: 0x3F, B: 0x43, A: 0xFF} // #F23F43
	colLEDYellow = color.NRGBA{R: 0xFE, G: 0xE7, B: 0x5C, A: 0xFF} // #FEE75C
	colLEDOrange = color.NRGBA{R: 0xFF, G: 0x99, B: 0x00, A: 0xFF} // #FF9900
)

type ledWidget struct {
	widget.BaseWidget
	state  LEDState
	circle *canvas.Circle
}

func newLEDWidget(initial LEDState) *ledWidget {
	w := &ledWidget{state: initial}
	w.circle = canvas.NewCircle(w.colorFor(initial))
	w.circle.Resize(fyne.NewSize(12, 12))
	w.ExtendBaseWidget(w)
	return w
}

func (w *ledWidget) colorFor(s LEDState) color.Color {
	switch s {
	case LEDGreen:
		return colLEDGreen
	case LEDRed:
		return colLEDRed
	case LEDYellow:
		return colLEDYellow
	case LEDOrange:
		return colLEDOrange
	default:
		return colLEDRed
	}
}

func (w *ledWidget) SetState(s LEDState) {
	w.state = s
	w.circle.FillColor = w.colorFor(s)
	w.circle.Refresh()
}

func (w *ledWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(w.circle)
}

func (w *ledWidget) MinSize() fyne.Size {
	return fyne.NewSize(14, 14)
}

// antiSpam debouncer for buttons
type antiSpam struct {
	mu     sync.Mutex
	locked bool
}

func (a *antiSpam) Run(fn func()) {
	a.mu.Lock()
	if a.locked {
		a.mu.Unlock()
		return
	}
	a.locked = true
	a.mu.Unlock()

	go func() {
		fn()
		time.Sleep(350 * time.Millisecond)
		a.mu.Lock()
		a.locked = false
		a.mu.Unlock()
	}()
}

// bottomToast represents a non-intrusive popup that spawns at the bottom of the window
type bottomToast struct {
	container *fyne.Container
	label     *widget.Label
	dismiss   *widget.Button
	timer     *time.Timer
	mu        sync.Mutex
}

func newBottomToast() *bottomToast {
	bt := &bottomToast{}
	bt.label = widget.NewLabel("")
	bt.label.TextStyle = fyne.TextStyle{Bold: true}

	bt.dismiss = widget.NewButton("Dismiss", func() {
		bt.Hide()
	})

	cardBg := canvas.NewRectangle(color.NRGBA{R: 0x11, G: 0x12, B: 0x14, A: 0xFA})
	cardBg.StrokeColor = color.NRGBA{R: 0x00, G: 0xA8, B: 0xFF, A: 0xC0}
	cardBg.StrokeWidth = 1.5
	cardBg.CornerRadius = 8

	content := container.NewHBox(
		layout.NewSpacer(),
		bt.label,
		bt.dismiss,
		layout.NewSpacer(),
	)

	padded := container.NewPadded(content)
	stack := container.NewStack(cardBg, padded)

	bt.container = container.NewVBox(
		layout.NewSpacer(),
		container.NewPadded(stack),
	)
	bt.container.Hide()

	return bt
}

func (bt *bottomToast) Show(msg string) {
	bt.mu.Lock()
	defer bt.mu.Unlock()

	if bt.timer != nil {
		bt.timer.Stop()
	}

	fyne.Do(func() {
		bt.label.SetText(msg)
		bt.container.Show()
		bt.container.Refresh()
	})

	bt.timer = time.AfterFunc(3500*time.Millisecond, func() {
		bt.Hide()
	})
}

func (bt *bottomToast) Hide() {
	bt.mu.Lock()
	defer bt.mu.Unlock()
	if bt.timer != nil {
		bt.timer.Stop()
	}
	fyne.Do(func() {
		bt.container.Hide()
		bt.container.Refresh()
	})
}

// signalLabel renders a clean signal indicator e.g. "85%"
func signalLabel(pct int) fyne.CanvasObject {
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	lbl := widget.NewLabel(fmt.Sprintf("%d%%", pct))
	lbl.TextStyle = fyne.TextStyle{Monospace: true}
	return lbl
}
