package main

import (
	"github.com/renato0307/terrui/internal/ui"
)

func main() {
	app := ui.NewApp()

	if err := app.Run(); err != nil {
		panic(err)
	}
}
