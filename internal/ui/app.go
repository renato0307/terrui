package ui

import (
	"fmt"
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/renato0307/terrui/internal/config"
	"github.com/rivo/tview"
)

const defaultFooter string = "💡press ? for help"

type App struct {
	*tview.Application

	layout      *tview.Grid
	pages       *tview.Pages
	pagesMap    map[string]PageFactory
	currentPage Page
	actions     KeyActions

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
	pages.SetBorderPadding(0, 0, 1, 1)
	layout := tview.NewGrid().
		SetRows(1, 0, 1).
		SetBorders(true)

	header := NewHeader().SetCrumb([]string{})
	footer := NewFooter(a, "welcome 🤓 - press ? for help", tview.Styles.PrimaryTextColor, 3)

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
	a.actions = a.bindKeys()

	if config.Organization == "" {
		a.activatePage(OrganizationsPageName, nil, false)
	} else {
		a.activatePage(WorkspacesPageName, nil, false)
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
	pagesMap[WorkspacesPageName] = NewWorkspacesPage
	pagesMap[WorkspacePageName] = NewWorkspacePage
	pagesMap[HelpPageName] = NewHelpPage

	return pagesMap
}

func (a *App) activatePage(name string, page Page, skipLoad bool) {
	if a.pages.HasPage(name) {
		a.pages.RemovePage(name)
	}

	if page == nil {
		pageFactory := a.pagesMap[name]
		page = pageFactory(a)
	}

	if page.Name() != HelpPageName {
		a.actions.Clear()
	}
	a.actions.Add(a.bindKeys())
	a.actions.Add(page.BindKeys())

	a.pages.AddAndSwitchToPage(name, page, true)

	if page.Footer() != "" {
		a.footer.ShowText(page.Footer())
	} else {
		a.footer.ShowText(defaultFooter)
	}

	a.currentPage = page

	go a.exec(page, skipLoad)
}

func (a *App) exec(p Page, skipLoad bool) {
	a.QueueUpdateDraw(func() {
		a.header.SetCrumb(p.Crumb())
		a.footer.Show("⏳loading...", tview.Styles.SecondaryTextColor)
	})
	a.QueueUpdateDraw(func() {
		if !skipLoad {
			err := p.Load()
			if err != nil {
				a.footer.ShowError(fmt.Sprintf("😵%s", err.Error()))
				return
			}
		}
		msg := p.View()

		if msg != "" {
			a.footer.Show(fmt.Sprintf("✅%s", msg), tview.Styles.SecondaryTextColor)
		} else {
			a.footer.ShowText(p.Footer())
		}
	})
}

func (a *App) appKeyboard(evt *tcell.EventKey) *tcell.EventKey {
	key := AsKey(evt)
	action, ok := a.actions[key]
	if ok {
		return action.Action(evt)
	}
	return evt
}

func (a *App) bindKeys() KeyActions {
	return KeyActions{
		tcell.KeyCtrlO: NewSharedKeyAction("list organizations", a.listOrgs, true),
		tcell.KeyCtrlC: NewSharedKeyAction("quit", a.quit, true),
		KeyHelp:        NewSharedKeyAction("help", a.showHelp, true),
	}
}

func (a *App) listOrgs(ek *tcell.EventKey) *tcell.EventKey {
	a.config.Organization = ""
	a.config.Save()
	a.activatePage(OrganizationsPageName, nil, false)

	return nil
}

func (a *App) quit(ek *tcell.EventKey) *tcell.EventKey {
	a.Stop()
	os.Exit(0)

	return nil
}

func (a *App) showHelp(ek *tcell.EventKey) *tcell.EventKey {
	currentPage := a.currentPage
	a.activatePage(HelpPageName, nil, false)
	a.currentPage = currentPage
	return nil
}

func AsKey(evt *tcell.EventKey) tcell.Key {
	if evt.Key() != tcell.KeyRune {
		return evt.Key()
	}
	key := tcell.Key(evt.Rune())
	if evt.Modifiers() == tcell.ModAlt {
		key = tcell.Key(int16(evt.Rune()) * int16(evt.Modifiers()))
	}
	return key
}
