package ui

import (
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type App struct {
	*tview.Application
	prompt *Prompt
	main   *tview.Grid
	header *tview.TextView
}

func NewApp() *App {
	newPrimitive := func(text string) *tview.TextView {
		t := tview.NewTextView().
			SetTextAlign(tview.AlignCenter).
			SetText(text)

		return t
	}

	grid := tview.NewGrid().
		SetRows(1, 0, 1).
		SetBorders(true)

	content := newPrimitive("")
	header := newPrimitive("terrui")
	prompt := NewPrompt(grid.GetBackgroundColor())

	grid.AddItem(header, 0, 0, 1, 3, 0, 0, false).
		AddItem(prompt, 2, 0, 1, 3, 0, 0, false).
		AddItem(content, 1, 0, 1, 3, 0, 0, false)

	app := tview.NewApplication()
	a := &App{
		Application: app,
		header:      header,
		main:        grid,
		prompt:      prompt,
	}
	prompt.AddListener("app", a)
	a.SetInputCapture(a.appKeyboard)

	return a
}

func (a *App) appKeyboard(evt *tcell.EventKey) *tcell.EventKey {
	// nolint:exhaustive
	a.header.SetText("pressed: \"" + string(evt.Rune()) + "\"")
	key := tcell.Key(evt.Rune())
	switch key {
	case KeyColon:
		a.prompt.Reset()
		a.SetFocus(a.prompt)
		return nil
	case tcell.KeyCtrlC:
		a.Stop()
		os.Exit(0)
	}

	return evt
}

func (a *App) Completed(text string) {
	a.header.SetText("cmd: " + text)
	a.SetFocus(a.main)
	a.SetInputCapture(a.appKeyboard)
}

func (a *App) Canceled() {
	a.SetFocus(a.main)
}

func (a *App) Run() error {
	return a.SetRoot(a.main, true).SetFocus(a.main).Run()
}
