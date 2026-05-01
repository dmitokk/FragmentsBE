package main

import (
	"log/slog"

	"github.com/dmitokk/FragmentsBE/internal/app"
)

func main() {
	application, err := app.New()
	if err != nil {
		slog.Error("Failed to initialize application", "error", err)
		return
	}

	if err := application.Run(); err != nil {
		slog.Error("Failed to run application", "error", err)
	}
}