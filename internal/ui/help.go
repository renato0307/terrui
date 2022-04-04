package ui

import (
	"sort"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const HelpPageName string = "help"

type HelpPage struct {
	*tview.Table

	app      *App
	returnTo Page
}

type helpEntry struct {
	action      string
	description string
}

func NewHelpPage(app *App) Page {
	h := &HelpPage{
		Table: tview.NewTable(),

		app:      app,
		returnTo: app.currentPage,
	}
	return h
}

func (h *HelpPage) Load() error {
	return nil
}

func (h *HelpPage) View() string {
	h.SetSelectable(true, false)
	h.SetCell(0, 0, tview.NewTableCell("KEY").SetSelectable(false))
	h.SetCell(0, 1, tview.NewTableCell("ACTION").SetSelectable(false))

	helpEntries := []helpEntry{}
	for k, a := range h.app.actions {
		helpEntries = append(helpEntries, helpEntry{
			action:      tcell.KeyNames[k],
			description: a.Description,
		})
	}
	sort.Slice(helpEntries, func(i, j int) bool { return helpEntries[i].description < helpEntries[j].description })

	for i, e := range helpEntries {
		r := i + 1
		h.SetCell(r, 0, tview.NewTableCell(e.action).SetExpansion(1))
		h.SetCell(r, 1, tview.NewTableCell(e.description).SetExpansion(2))
	}

	return ""
}

func (h *HelpPage) BindKeys() KeyActions {
	return KeyActions{
		tcell.KeyEsc: NewKeyAction("exit help", h.exitHelp, true),
	}
}

func (h *HelpPage) Crumb() []string {
	return []string{"help"}
}

func (h *HelpPage) exitHelp(ek *tcell.EventKey) *tcell.EventKey {
	h.app.activatePage(h.app.currentPage.Name(), h.app.currentPage, true)
	return nil
}

func (h *HelpPage) Name() string {
	return HelpPageName
}

func (h *HelpPage) Footer() string {
	return "💡press <esc> to go back"
}
