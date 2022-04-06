package ui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/hashicorp/go-tfe"
	"github.com/rivo/tview"

	"github.com/renato0307/terrui/internal/client"
)

const WorkspacesPageName string = "workspaces"

type WorkspacesPage struct {
	*tview.Flex
	table      *tview.Table
	pagination *tview.TextView

	app              *App
	workspaces       *tfe.WorkspaceList
	currentWorkspace int
}

func NewWorkspacesPage(app *App) Page {
	w := WorkspacesPage{
		Flex:             tview.NewFlex(),
		table:            tview.NewTable(),
		pagination:       tview.NewTextView(),
		app:              app,
		currentWorkspace: 1,
	}

	w.pagination.SetTextAlign(tview.AlignRight)

	w.AddItem(w.pagination, 2, 0, false).
		SetDirection(tview.FlexRow).
		AddItem(w.table, 0, 1, true)

	return &w
}

func (w *WorkspacesPage) Load() error {
	return w.load(-1)
}

func (w *WorkspacesPage) load(pageNumber int) error {
	tfeClient, err := client.NewTFEClient()
	if err != nil {
		return fmt.Errorf("error creating the TFE client: %w", err)
	}

	workspaces, err := tfeClient.ListWorkspaces(w.app.config.Organization, pageNumber)
	if err != nil {
		return fmt.Errorf("error listing the organization: %w", err)
	}
	w.workspaces = workspaces

	return nil
}

func (w *WorkspacesPage) View() string {
	w.table.SetSelectable(true, false)
	w.table.SetCell(0, 0, tview.NewTableCell("ID").SetSelectable(false))
	w.table.SetCell(0, 1, tview.NewTableCell("NAME").SetSelectable(false))
	w.table.SetCell(0, 2, tview.NewTableCell("TAGS").SetSelectable(false))
	w.table.SetCell(0, 3, tview.NewTableCell("TERRAFORM").SetSelectable(false))
	w.table.SetCell(0, 4, tview.NewTableCell("COUNT").SetSelectable(false))
	w.table.SetCell(0, 5, tview.NewTableCell("RUN STATUS").SetSelectable(false))
	w.table.SetCell(0, 5, tview.NewTableCell("LATEST CHANGE").SetSelectable(false))

	if w.workspaces == nil || w.workspaces.TotalCount == 0 {
		return "no workspaces found"
	}

	for i, wi := range w.workspaces.Items {
		r := i + 1
		w.table.SetCell(r, 0, tview.NewTableCell(wi.ID).SetExpansion(1))
		w.table.SetCell(r, 1, tview.NewTableCell(wi.Name).SetExpansion(1))
		w.table.SetCell(r, 2, fmtTags(wi).SetExpansion(1))
		w.table.SetCell(r, 3, tview.NewTableCell(wi.TerraformVersion).SetExpansion(2))
		w.table.SetCell(r, 4, tview.NewTableCell(fmt.Sprint(wi.ResourceCount)).SetExpansion(2))
		w.table.SetCell(r, 5, fmtCurrentRun(wi).SetExpansion(1))
		w.table.SetCell(r, 6, fmtUpdatedAt(wi).SetExpansion(1))
	}

	w.table.SetSelectionChangedFunc(func(row, column int) {
		w.currentWorkspace = row
	})

	w.pagination.SetText(
		fmt.Sprintf("page %d of %d, total workspaces: %d",
			w.workspaces.CurrentPage,
			w.workspaces.TotalPages,
			w.workspaces.TotalCount))

	return "workspaces loaded"
}

func fmtTags(w *tfe.Workspace) *tview.TableCell {
	s := fmt.Sprint(w.TagNames)
	return tview.NewTableCell(strings.Trim(s, "[]"))
}

func fmtUpdatedAt(w *tfe.Workspace) *tview.TableCell {
	return tview.NewTableCell(w.UpdatedAt.Local().Format("02 Jan 06 15:04 MST"))
}

func fmtCurrentRun(w *tfe.Workspace) *tview.TableCell {
	if w.CurrentRun == nil {
		return tview.NewTableCell("")
	}

	style := tcell.StyleDefault
	style = style.Background(tview.Styles.PrimitiveBackgroundColor)
	switch w.CurrentRun.Status {
	case tfe.RunErrored:
		style = style.Bold(true)
		style = style.Foreground(tcell.ColorRed)
	case tfe.RunApplied, tfe.RunPlannedAndFinished:
		style = style.Bold(true)
		style = style.Foreground(tcell.ColorGreen)
	case tfe.RunPlanned, tfe.RunPlanning:
		style = style.Bold(true)
		style = style.Foreground(tcell.ColorYellow)
	default:
		style = style.Foreground(tview.Styles.PrimaryTextColor)
	}

	return tview.NewTableCell(fmt.Sprint(w.CurrentRun.Status)).SetStyle(style)
}

func (w *WorkspacesPage) BindKeys() KeyActions {
	return KeyActions{
		tcell.KeyEnter: NewKeyAction("select workspace", w.actionSelectWorkspace, true),
		tcell.KeyCtrlJ: NewKeyAction("next page", w.actionPaginationNextPage, true),
		tcell.KeyCtrlK: NewKeyAction("previous page", w.actionPaginationPrevPage, true),
	}

}

func (w *WorkspacesPage) Crumb() []string {
	return []string{
		w.app.config.Organization,
		WorkspacesPageName,
	}
}

func (w *WorkspacesPage) Name() string {
	return WorkspacesPageName
}

func (w *WorkspacesPage) Footer() string {
	return ""
}

func (w *WorkspacesPage) actionSelectWorkspace(ek *tcell.EventKey) *tcell.EventKey {
	w.app.config.Workspace = w.table.GetCell(w.currentWorkspace, 1).Text
	w.app.config.Save()

	w.app.activatePage(WorkspacePageName, nil, false)

	return nil
}

func (w *WorkspacesPage) actionPaginationNextPage(ek *tcell.EventKey) *tcell.EventKey {
	go w.app.ExecPageWithLoadFunc(w, w.loadNextPageFunc(), false)
	return nil
}

func (w *WorkspacesPage) actionPaginationPrevPage(ek *tcell.EventKey) *tcell.EventKey {
	go w.app.ExecPageWithLoadFunc(w, w.loadPrevPageFunc(), false)
	return nil
}

func (w *WorkspacesPage) loadNextPageFunc() func() error {
	return func() error {
		pageToLoad := w.workspaces.CurrentPage + 1
		if pageToLoad > w.workspaces.TotalPages {
			pageToLoad = 1
		}
		return w.load(pageToLoad)
	}
}

func (w *WorkspacesPage) loadPrevPageFunc() func() error {
	return func() error {
		pageToLoad := w.workspaces.CurrentPage - 1
		if pageToLoad < 1 {
			pageToLoad = w.workspaces.TotalPages
		}
		return w.load(pageToLoad)
	}
}
