package ui

import (
	"fmt"
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type App struct {
	*tview.Application
	layout  *tview.Grid
	content *tview.Pages
	prompt  *Prompt
	header  *Header
	message *Header
}

func NewApp() *App {
	app := tview.NewApplication()
	a := &App{}

	content := tview.NewPages()
	title := NewHeader(a, "terrUI", tcell.ColorWhiteSmoke, 0)
	message := NewHeader(a, "$ Welcome - press \":\" for commands or \"?\" for help", tcell.ColorYellow, 3)

	grid := tview.NewGrid().
		SetRows(1, 1, 0, 1).
		SetBorders(true)

	prompt := NewPrompt(nil, grid.GetBackgroundColor())
	prompt.AddListener("app", a)
	prompt.SetApp(a)

	grid.AddItem(title, 0, 0, 1, 2, 0, 0, false).
		AddItem(message, 1, 0, 1, 2, 0, 0, false).
		AddItem(tview.NewTextView().SetText(" organization: - "), 0, 2, 1, 2, 0, 0, false).
		AddItem(tview.NewTextView().SetText(" workspace: - "), 1, 2, 1, 2, 0, 0, false).
		AddItem(content, 2, 0, 1, 4, 0, 0, false).
		AddItem(prompt, 3, 0, 1, 4, 0, 0, false)

	a.Application = app
	a.content = content
	a.header = title
	a.message = message
	a.layout = grid
	a.prompt = prompt
	a.SetInputCapture(a.appKeyboard)

	return a
}

func (a *App) appKeyboard(evt *tcell.EventKey) *tcell.EventKey {
	// nolint:exhaustive
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
		a.message.ShowText("> organizations")
		orgList, err := NewOrganizationList(a)
		if err != nil {
			a.message.ShowError("could not connect to TFE")
			return
		}
		a.content.AddAndSwitchToPage("orgs", orgList, true)
		go orgList.Load()
		return
	}

	a.message.ShowError(fmt.Sprintf("invalid command: %s", text))
}

func (a *App) Canceled() {
	a.ResetFocus()
}

func (a *App) ResetFocus() {
	_, frontPage := a.content.GetFrontPage()
	a.SetFocus(frontPage)
}

func (a *App) Run() error {
	return a.SetRoot(a.layout, true).SetFocus(a.layout).Run()
}
