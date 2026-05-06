package main

import (
	"embed"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"skill-manager/internal/cli"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Dispatch CLI subcommands before starting the Wails window.
	if len(os.Args) > 1 && os.Args[1] == "skills" {
		os.Exit(cli.Run(os.Args[2:]))
	}

	app := NewApp()

	err := wails.Run(&options.App{
		Title:  "skill-manager",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
