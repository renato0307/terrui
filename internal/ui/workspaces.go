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

type WorkspacesPageSource struct {
	app        *App
	workspaces *tfe.WorkspaceList
}

func NewWorkspacesPage(app *App) Page {
	return NewListPage(app, &WorkspacesPageSource{app: app})
}

func (w *WorkspacesPageSource) SupportsSearch() bool {
	return true
}

func (w *WorkspacesPageSource) Search(searchText string, pageNumber int) error {
	tfeClient, err := client.NewTFEClient()
	if err != nil {
		return fmt.Errorf("error creating the TFE client: %w", err)
	}

	workspaces, err := tfeClient.ListWorkspaces(w.app.config.Organization, searchText, pageNumber)
	if err != nil {
		return fmt.Errorf("error listing the workspaces: %w", err)
	}
	w.workspaces = workspaces

	return nil
}

func (w *WorkspacesPageSource) RenderHeader(table *tview.Table) {
	table.SetCell(0, 0, tview.NewTableCell("ID").SetSelectable(false))
	table.SetCell(0, 1, tview.NewTableCell("NAME").SetSelectable(false))
	table.SetCell(0, 2, tview.NewTableCell("TAGS").SetSelectable(false))
	table.SetCell(0, 3, tview.NewTableCell("TERRAFORM").SetSelectable(false))
	table.SetCell(0, 4, tview.NewTableCell("COUNT").SetSelectable(false))
	table.SetCell(0, 5, tview.NewTableCell("RUN STATUS").SetSelectable(false))
	table.SetCell(0, 5, tview.NewTableCell("LATEST CHANGE").SetSelectable(false))
}

func (w *WorkspacesPageSource) RenderRows(table *tview.Table) {
	for i, wi := range w.workspaces.Items {
		r := i + 1
		table.SetCell(r, 0, tview.NewTableCell(wi.ID).SetExpansion(1))
		table.SetCell(r, 1, tview.NewTableCell(wi.Name).SetExpansion(1))
		table.SetCell(r, 2, fmtTags(wi).SetExpansion(1))
		table.SetCell(r, 3, tview.NewTableCell(wi.TerraformVersion).SetExpansion(2))
		table.SetCell(r, 4, tview.NewTableCell(fmt.Sprint(wi.ResourceCount)).SetExpansion(2))
		table.SetCell(r, 5, fmtCurrentRun(wi).SetExpansion(1))
		table.SetCell(r, 6, fmtUpdatedAt(wi).SetExpansion(1))
	}
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

func (w *WorkspacesPageSource) Crumb() []string {
	return []string{
		w.app.config.Organization,
		WorkspacesPageName,
	}
}

func (w *WorkspacesPageSource) Name() string {
	return "workspace"
}

func (w *WorkspacesPageSource) NameList() string {
	return WorkspacesPageName
}

func (w *WorkspacesPageSource) ActionSelectWorkspace(table *tview.Table, currentItem int) func(ek *tcell.EventKey) *tcell.EventKey {
	return func(ek *tcell.EventKey) *tcell.EventKey {
		w.app.config.Workspace = table.GetCell(currentItem, 1).Text
		w.app.config.Save()
		w.app.activatePage(WorkspacePageName, nil, false)
		return nil
	}
}

func (w *WorkspacesPageSource) Empty() bool {
	return w.workspaces == nil || len(w.workspaces.Items) == 0
}

func (w *WorkspacesPageSource) CurrentPage() int {
	return w.workspaces.CurrentPage
}

func (w *WorkspacesPageSource) TotalCount() int {
	return w.workspaces.TotalCount
}

func (w *WorkspacesPageSource) TotalPages() int {
	return w.workspaces.TotalPages
}
