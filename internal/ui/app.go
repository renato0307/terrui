package ui

import (
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type App struct {
	*tview.Application
	layout  *tview.Grid
	content *tview.Pages
	prompt  *Prompt
	header  *tview.TextView
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

	content := tview.NewPages()
	header := newPrimitive("terrui")
	prompt := NewPrompt(nil, grid.GetBackgroundColor())

	grid.AddItem(header, 0, 0, 1, 3, 0, 0, false).
		AddItem(prompt, 2, 0, 1, 3, 0, 0, false).
		AddItem(content, 1, 0, 1, 3, 0, 0, false)

	app := tview.NewApplication()
	a := &App{
		Application: app,
		content:     content,
		header:      header,
		layout:      grid,
		prompt:      prompt,
	}
	prompt.AddListener("app", a)
	prompt.SetApp(a)
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
	a.ResetFocus()
	if text == "orgs" || text == "o" {
		a.header.SetText("organizations")
		orgList := NewOrganizationList(a)
		a.content.AddAndSwitchToPage("orgs", orgList, true)
		go orgList.Load()
		return
	}

	a.header.SetText("invalid command")
}

func (a *App) Canceled() {
	_, frontPage := a.content.GetFrontPage()
	a.SetFocus(frontPage)
}

func (a *App) ResetFocus() {
	a.SetFocus(a.layout)
}

func (a *App) Run() error {
	return a.SetRoot(a.layout, true).SetFocus(a.layout).Run()
}
