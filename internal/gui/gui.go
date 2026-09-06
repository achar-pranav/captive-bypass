package gui

import (
	"context"
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"

	"github.com/achar-pranav/captive-bypass/backends/auto"
	"github.com/achar-pranav/captive-bypass/internal/config"
	"github.com/achar-pranav/captive-bypass/internal/portal"
)

//go:embed all:frontend/dist
var assets embed.FS

// Run initializes and launches the Wails AMOLED control panel.
func Run() error {
	dir := config.DefaultDir()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}
	cfgPath := filepath.Join(dir, "config.json")
	cfg, err := config.Load(cfgPath)
	if err != nil && err != config.ErrNoConfig {
		return fmt.Errorf("loading config: %w", err)
	}

	app := NewApp(cfg, cfgPath, portal.New("", nil), auto.Default())

	return wails.Run(&options.App{
		Title:             "captive-bypass",
		Width:             460,
		Height:            720,
		MinWidth:          420,
		MinHeight:         640,
		DisableResize:     false,
		Fullscreen:        false,
		Frameless:         false,
		StartHidden:       false,
		HideWindowOnClose: false,
		BackgroundColour:  &options.RGBA{R: 0, G: 0, B: 0, A: 255},
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: func(ctx context.Context) {
			app.Startup(ctx)
		},
		Bind: []interface{}{
			app,
		},
		Linux: &linux.Options{
			WindowIsTranslucent: false,
			WebviewGpuPolicy:    linux.WebviewGpuPolicyAlways,
		},
		Mac: &mac.Options{
			TitleBar:             mac.TitleBarHiddenInset(),
			Appearance:           mac.NSAppearanceNameDarkAqua,
			WebviewIsTransparent: false,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			BackdropType:         windows.None,
			Theme:                windows.Dark,
		},
	})
}
