package ui

import (
	"fmt"
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/renato0307/terrui/internal/config"
	"github.com/rivo/tview"
)

type App struct {
	*tview.Application

	layout   *tview.Grid
	pages    *tview.Pages
	pagesMap map[string]PageFactory
	actions  KeyActions

	header *Header
	footer *Footer

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
	layout := tview.NewGrid().
		SetRows(1, 0, 1).
		SetBorders(true)

	header := NewHeader()
	footer := NewFooter(a, "welcome 🤓 - press ? for help", tview.Styles.PrimaryTextColor, 5)

	layout.AddItem(header, 0, 0, 1, 1, 0, 0, false).
		AddItem(pages, 1, 0, 1, 1, 0, 0, false).
		AddItem(footer, 2, 0, 1, 1, 0, 0, false)

	a.Application = app
	a.config = config
	a.layout = layout
	a.pages = pages
	a.header = header
	a.footer = footer
	a.pagesMap = initPages()
	a.actions = KeyActions{}

	if config.Organization == "" {
		a.activatePage(OrganizationsPageName)
	} else {
		// a.gotoOrganization()
	}

	a.SetInputCapture(a.appKeyboard)
	return a
}

func (a *App) Run() error {
	return a.SetRoot(a.layout, true).SetFocus(a.pages).Run()
}

func initPages() map[string]PageFactory {
	pagesMap := map[string]PageFactory{}

	pagesMap[OrganizationsPageName] = NewOrganizationsPage
	pagesMap[OrganizationPageName] = NewOrganizationPage

	return pagesMap
}

func (a *App) activatePage(name string) {
	a.header.GoForward(name)
	if a.pages.HasPage(name) {
		a.pages.RemovePage(name)
	}

	pageFactory := a.pagesMap[name]
	page := pageFactory(a)
	a.actions.Add(page.BindKeys())
	a.pages.AddAndSwitchToPage(name, page, true)
	a.footer.ShowText("press ? for help")

	go a.exec(page)
}

func (a *App) exec(p Page) {
	a.QueueUpdateDraw(func() {
		a.footer.Show("⏳ loading...", tview.Styles.SecondaryTextColor)
	})
	a.QueueUpdateDraw(func() {
		err := p.Load()
		if err != nil {
			a.footer.ShowError(err.Error())
			return
		}
		msg := p.View()
		a.footer.Show(fmt.Sprintf("✅ %s", msg), tview.Styles.SecondaryTextColor)
	})
}

func (a *App) appKeyboard(evt *tcell.EventKey) *tcell.EventKey {
	// nolint:exhaustive
	key := tcell.Key(evt.Rune())
	action, ok := a.actions[key]
	if ok {
		return action.Action(evt)
	}

	switch key {
	case tcell.KeyCtrlC:
		a.Stop()
		os.Exit(0)
	}

	return evt
}
