package ui

import (
	"fmt"
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/renato0307/terrui/legacy/config"
	"github.com/rivo/tview"
)

type App struct {
	*tview.Application

	layout *tview.Grid
	pages  *tview.Pages

	header       *Header
	message      *Header
	workspace    *Header
	organization *Header

	supportedCmds *SupportedCommands
	prompt        *Prompt

	config *config.Config
}

func NewApp() *App {
	app := tview.NewApplication()
	a := &App{}

	config, err := config.NewConfig()
	if err != nil {
		panic(err)
	}

	pages := tview.NewPages()
	title := NewHeader(a, "terrUI", "", tview.Styles.PrimaryTextColor, 0)
	message := NewHeader(a, "$ Welcome 🤓", "", tcell.ColorYellow, 3)

	layout := tview.NewGrid().
		SetRows(1, 1, 0, 4, 1).
		SetBorders(true)

	prompt := NewPrompt(nil, layout.GetBackgroundColor())
	prompt.AddListener("app", a)
	prompt.SetApp(a)

	scList := []SupportedCommand{}
	supportedCommands := NewSupportedCommands(a, scList, 3)

	workspace := NewHeader(a, config.Workspace, "workspace", tview.Styles.PrimaryTextColor, 0)

	organization := NewHeader(a, config.Organization, "organization", tview.Styles.PrimaryTextColor, 0)

	layout.AddItem(title, 0, 0, 1, 2, 0, 0, false).
		AddItem(message, 1, 0, 1, 2, 0, 0, false).
		AddItem(organization, 0, 2, 1, 2, 0, 0, false).
		AddItem(workspace, 1, 2, 1, 2, 0, 0, false).
		AddItem(pages, 2, 0, 1, 4, 0, 0, false).
		AddItem(supportedCommands, 3, 0, 1, 4, 0, 0, false).
		AddItem(prompt, 4, 0, 1, 4, 0, 0, false)

	a.Application = app
	a.config = config
	a.header = title
	a.layout = layout
	a.message = message
	a.organization = organization
	a.pages = pages
	a.prompt = prompt
	a.supportedCmds = supportedCommands
	a.workspace = workspace

	a.SetInputCapture(a.appKeyboard)

	return a
}

func (a *App) Run() error {
	if a.config.Organization != "" && a.config.Workspace != "" {
		a.showWorkspace()
	}

	return a.SetRoot(a.layout, true).SetFocus(a.layout).Run()
}

func (a *App) appKeyboard(evt *tcell.EventKey) *tcell.EventKey {
	// nolint:exhaustive
	key := tcell.Key(evt.Rune())
	switch key {
	case tcell.KeyCtrlC:
		a.Stop()
		os.Exit(0)
	}

	return evt
}

// for prompt
func (a *App) Completed(text string) {
	a.ResetFocus()
	a.processCommand(text)
}

func (a *App) processCommand(text string) {
	switch text {
	case "orgs", "o":
		a.showOrganizationList()
	case "workspaces", "w":
		a.showWorkspaceList()
	default:
		a.message.ShowError(fmt.Sprintf("invalid command: %s", text))
	}
}

func (a *App) Canceled() {
	a.ResetFocus()
}

func (a *App) ResetFocus() {
	_, frontPage := a.pages.GetFrontPage()
	a.SetFocus(frontPage)
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

func (a *App) showWorkspaceList() error {
	wl, err := NewWorkspaceList(a, a.config.Organization)
	if err != nil {
		return err
	}
	a.pages.AddAndSwitchToPage("workspaces", wl, true)
	go wl.Load()

	return nil
}

func (a *App) showWorkspace() error {
	w, err := NewWorkspace(a)
	if err != nil {
		return err
	}
	a.pages.AddAndSwitchToPage("workspace", w, true)
	go w.Load()

	return nil
}
