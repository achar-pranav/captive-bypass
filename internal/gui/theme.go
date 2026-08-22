package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// Electric blue on AMOLED black — Discord card layout, recolored.
var (
	colBlack    = color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xFF}
	colCard     = color.NRGBA{R: 0x0D, G: 0x0F, B: 0x12, A: 0xFF}
	colCardEdge = color.NRGBA{R: 0x00, G: 0xA8, B: 0xFF, A: 0x2E}
	colBlue     = color.NRGBA{R: 0x00, G: 0xA8, B: 0xFF, A: 0xFF}
	colBlueDim  = color.NRGBA{R: 0x00, G: 0x5A, B: 0x87, A: 0xFF}
	colText     = color.NRGBA{R: 0xE6, G: 0xEA, B: 0xED, A: 0xFF}
	colTextDim  = color.NRGBA{R: 0x7A, G: 0x82, B: 0x8A, A: 0xFF}
	colInputBg  = color.NRGBA{R: 0x13, G: 0x16, B: 0x1A, A: 0xFF}
)

type amoledTheme struct {
	fyne.Theme
}

func newAmoledTheme() fyne.Theme {
	return &amoledTheme{Theme: fyne.CurrentApp().Settings().Theme()}
}

func (t *amoledTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return colBlack
	case theme.ColorNameButton:
		return colCard
	case theme.ColorNameInputBackground:
		return colInputBg
	case theme.ColorNamePrimary:
		return colBlue
	case theme.ColorNameFocus, theme.ColorNameHover, theme.ColorNamePressed, theme.ColorNameSelection:
		return colBlueDim
	case theme.ColorNameForegroundOnPrimary:
		return color.NRGBA{R: 0x00, G: 0x08, B: 0x10, A: 0xFF}
	case theme.ColorNameSeparator:
		return color.NRGBA{R: 0x1C, G: 0x22, B: 0x28, A: 0xFF}
	case theme.ColorNameScrollBar:
		return colBlueDim
	case theme.ColorNamePlaceHolder:
		return colTextDim
	default:
		return t.Theme.Color(name, variant)
	}
}

func (t *amoledTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameInputRadius, theme.SizeNameSelectionRadius:
		return 8
	case theme.SizeNameInnerWindowRadius:
		return 10
	default:
		return t.Theme.Size(name)
	}
}
