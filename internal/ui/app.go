package ui

import (
	"fmt"
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type App struct {
	*tview.Application
	layout        *tview.Grid
	pages         *tview.Pages
	prompt        *Prompt
	header        *Header
	message       *Header
	supportedCmds *SupportedCommands
}

func NewApp() *App {
	app := tview.NewApplication()
	a := &App{}

	pages := tview.NewPages()
	title := NewHeader(a, "terrUI", tcell.ColorWhiteSmoke, 0)
	message := NewHeader(a, "$ Welcome 🤓", tcell.ColorYellow, 3)

	layout := tview.NewGrid().
		SetRows(1, 1, 0, 4, 1).
		SetBorders(true)

	prompt := NewPrompt(nil, layout.GetBackgroundColor())
	prompt.AddListener("app", a)
	prompt.SetApp(a)

	scList := []SupportedCommand{
		// {
		// 	ShortCut: "esc",
		// 	Name:     "go back",
		// },
		{
			ShortCut: "ctrl+c",
			Name:     "quit",
		},
		// {
		// 	ShortCut: "?",
		// 	Name:     "show help",
		// },
	}
	supportedCommands := NewSupportedCommands(a, scList, 3)

	layout.AddItem(title, 0, 0, 1, 2, 0, 0, false).
		AddItem(message, 1, 0, 1, 2, 0, 0, false).
		AddItem(tview.NewTextView().SetText(" organization: - "), 0, 2, 1, 2, 0, 0, false).
		AddItem(tview.NewTextView().SetText(" workspace: - "), 1, 2, 1, 2, 0, 0, false).
		AddItem(pages, 2, 0, 1, 4, 0, 0, false).
		AddItem(supportedCommands, 3, 0, 1, 4, 0, 0, false).
		AddItem(prompt, 4, 0, 1, 4, 0, 0, false)

	a.Application = app
	a.pages = pages
	a.header = title
	a.message = message
	a.layout = layout
	a.supportedCmds = supportedCommands
	a.prompt = prompt
	a.SetInputCapture(a.appKeyboard)

	return a
}

func (a *App) Run() error {
	return a.SetRoot(a.layout, true).SetFocus(a.layout).Run()
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
	a.processCommand(text)
}

func (a *App) Canceled() {
	a.ResetFocus()
}

func (a *App) ResetFocus() {
	_, frontPage := a.pages.GetFrontPage()
	a.SetFocus(frontPage)
}

func (a *App) processCommand(text string) {
	switch text {
	case "orgs", "o":
		a.showOrganizationList()
	default:
		a.message.ShowError(fmt.Sprintf("invalid command: %s", text))
	}
}

func (a *App) showOrganizationList() {
	a.message.ShowText("> organizations")
	orgList, err := NewOrganizationList(a)
	if err != nil {
		return
	}
	a.pages.AddAndSwitchToPage("orgs", orgList, true)
	go orgList.Load()
}
