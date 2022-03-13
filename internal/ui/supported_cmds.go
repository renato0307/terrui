package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type SupportedCommand struct {
	ShortCut string
	Name     string
}

type SupportedCommands struct {
	*tview.Grid

	app     *App
	perLine int
}

func NewSupportedCommands(app *App, initialList []SupportedCommand, perLine int) *SupportedCommands {
	sc := SupportedCommands{
		perLine: perLine,
		app:     app,
	}

	sc.SetCommands(initialList)
	return &sc
}

func (sc *SupportedCommands) SetCommands(list []SupportedCommand) {
	sc.Grid = tview.NewGrid()

	sc.SetRows(1, 0).SetBorders(false)

	title := tview.NewTextView().SetText("available shortcuts:").SetTextColor(tcell.ColorCornflowerBlue)
	table := tview.NewTable()
	table.SetSelectable(false, false)
	table.SetBorder(false)
	for i, c := range list {
		cmdText := fmt.Sprintf("(%s) %s", c.ShortCut, c.Name)
		table.SetCell(i%sc.perLine, i/sc.perLine, tview.NewTableCell(cmdText).SetExpansion(1))
	}

	sc.AddItem(title, 0, 0, 1, 1, 0, 0, false)
	sc.AddItem(table, 1, 0, 1, 1, 0, 0, false)
}
