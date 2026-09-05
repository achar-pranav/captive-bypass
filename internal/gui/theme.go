package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// Discord AMOLED Theme
// Inspired by Discord's Midnight / AMOLED palette:
// Pure black background (#000000) with deep slate card surfaces (#111214 / #1E1F22)
// and Discord Blurple (#5865F2) accents.

var (
	colAmoledBg       = color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xFF} // #000000
	colDiscordCard    = color.NRGBA{R: 0x11, G: 0x12, B: 0x14, A: 0xFF} // #111214
	colDiscordSurface = color.NRGBA{R: 0x1E, G: 0x1F, B: 0x22, A: 0xFF} // #1E1F22
	colDiscordBorder  = color.NRGBA{R: 0x2B, G: 0x2D, B: 0x31, A: 0xFF} // #2B2D31
	colBlurple        = color.NRGBA{R: 0x58, G: 0x65, B: 0xF2, A: 0xFF} // #5865F2
	colBlurpleHover   = color.NRGBA{R: 0x47, G: 0x52, B: 0xC4, A: 0xFF} // #4752C4
	colBlurpleActive  = color.NRGBA{R: 0x3C, G: 0x45, B: 0xA5, A: 0xFF} // #3C45A5
	colTextLight      = color.NRGBA{R: 0xF2, G: 0xF3, B: 0xF5, A: 0xFF} // #F2F3F5
	colTextMuted      = color.NRGBA{R: 0x94, G: 0x9B, B: 0xA4, A: 0xFF} // #949BA4
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
		return colDiscordSurface
	case theme.ColorNameInputBackground:
		return colDiscordSurface
	case theme.ColorNameMenuBackground, theme.ColorNameOverlayBackground:
		return colDiscordCard
	case theme.ColorNamePrimary:
		return colBlurple
	case theme.ColorNameFocus, theme.ColorNameHover, theme.ColorNameSelection:
		return colBlurpleHover
	case theme.ColorNamePressed:
		return colBlurpleActive
	case theme.ColorNameForeground:
		return colTextLight
	case theme.ColorNameForegroundOnPrimary:
		return color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}
	case theme.ColorNameSeparator:
		return colDiscordBorder
	case theme.ColorNameScrollBar:
		return colDiscordBorder
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
	case theme.SizeNameInputRadius, theme.SizeNameSelectionRadius:
		return 8
	case theme.SizeNameInnerWindowRadius:
		return 10
	case theme.SizeNamePadding:
		return 8
	default:
		return t.Theme.Size(name)
	}
}
