package main

import (
	"embed"
	"log/slog"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	app := NewApp(logger)
	defer app.Close()

	if err := wails.Run(&options.App{
		Title:     "WaveArchive",
		Width:     1440,
		Height:    900,
		MinWidth:  1100,
		MinHeight: 680,
		AssetServer: &assetserver.Options{
			Assets:  assets,
			Handler: app.assetServerHandler(),
		},
		BackgroundColour: &options.RGBA{R: 8, G: 12, B: 18, A: 1},
		OnStartup:        app.Startup,
		Bind:             []interface{}{app},
	}); err != nil {
		logger.Error("application stopped", "error", err)
	}
}
