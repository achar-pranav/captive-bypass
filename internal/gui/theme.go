package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

var (
	// AMOLED pure pitch black & dark surfaces
	colAmoledBg       = color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xFF} // #000000
	colSurfaceCard    = color.NRGBA{R: 0x0A, G: 0x0C, B: 0x0E, A: 0xFF} // #0A0C0E
	colInputSurface   = color.NRGBA{R: 0x0E, G: 0x10, B: 0x12, A: 0xFF} // #0E1012
	colBorderDim      = color.NRGBA{R: 0x1E, G: 0x1F, B: 0x22, A: 0xFF} // #1E1F22
	colButtonBlack    = color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xFF} // #000000 with white outline

	// Electric Blue accents
	colElectricBlue   = color.NRGBA{R: 0x00, G: 0xA8, B: 0xFF, A: 0xFF} // #00A8FF
	colElectricHover  = color.NRGBA{R: 0x33, G: 0xBA, B: 0xFF, A: 0xFF} // #33BAFF
	colElectricActive = color.NRGBA{R: 0x00, G: 0x90, B: 0xDC, A: 0xFF} // #0090DC

	// Hover subtle white tint
	colHoverWhiteTint = color.NRGBA{R: 0x26, G: 0x2A, B: 0x30, A: 0xFF} // subtle white tint over dark

	// Typography & status
	colTextLight      = color.NRGBA{R: 0xF2, G: 0xF3, B: 0xF5, A: 0xFF} // #F2F3F5
	colTextMuted      = color.NRGBA{R: 0x7A, G: 0x82, B: 0x8A, A: 0xFF} // #7A828A
	colSuccess        = color.NRGBA{R: 0x23, G: 0xA5, B: 0x5A, A: 0xFF} // #23A55A
	colDanger         = color.NRGBA{R: 0xF2, G: 0x3F, B: 0x43, A: 0xFF} // #F23F43
)

type amoledTheme struct {
	fyne.Theme
}

func newAmoledTheme() fyne.Theme {
	return &amoledTheme{Theme: theme.DefaultTheme()}
}

func (t *amoledTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return colAmoledBg
	case theme.ColorNameButton:
		return colButtonBlack
	case theme.ColorNameInputBackground:
		return colInputSurface
	case theme.ColorNameMenuBackground, theme.ColorNameOverlayBackground:
		return colSurfaceCard
	case theme.ColorNamePrimary:
		return colElectricBlue
	case theme.ColorNameFocus, theme.ColorNameSelection:
		return colElectricHover
	case theme.ColorNameHover:
		return colHoverWhiteTint
	case theme.ColorNamePressed:
		return colElectricActive
	case theme.ColorNameForeground:
		return colTextLight
	case theme.ColorNameForegroundOnPrimary:
		return color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xFF}
	case theme.ColorNameSeparator:
		return colBorderDim
	case theme.ColorNameScrollBar:
		return colBorderDim
	case theme.ColorNamePlaceHolder:
		return colTextMuted
	case theme.ColorNameSuccess:
		return colSuccess
	case theme.ColorNameError:
		return colDanger
	default:
		return t.Theme.Color(name, theme.VariantDark)
	}
}

func (t *amoledTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameInputBorder:
		return 1
	case theme.SizeNameSelectionRadius:
		return 8
	case theme.SizeNameInnerWindowRadius:
		return 10
	case theme.SizeNamePadding:
		return 8
	default:
		return t.Theme.Size(name)
	}
}
