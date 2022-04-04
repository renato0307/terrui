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
	*tview.Table

	app              *App
	workspaces       *tfe.WorkspaceList
	currentWorkspace int
}

func NewWorkspacesPage(app *App) Page {
	ol := WorkspacesPage{
		Table: tview.NewTable(),
		app:   app,

		currentWorkspace: 1,
	}

	return &ol
}

func (w *WorkspacesPage) Load() error {

	tfeClient, err := client.NewTFEClient()
	if err != nil {
		return fmt.Errorf("error creating the TFE client: %w", err)
	}

	workspaces, err := tfeClient.ListWorkspaces(w.app.config.Organization)
	if err != nil {
		return fmt.Errorf("error listing the organization: %w", err)
	}
	w.workspaces = workspaces

	return nil
}

func (wl *WorkspacesPage) View() string {
	wl.Table.SetSelectable(true, false)
	wl.Table.SetCell(0, 0, tview.NewTableCell("ID").SetSelectable(false))
	wl.Table.SetCell(0, 1, tview.NewTableCell("NAME").SetSelectable(false))
	wl.Table.SetCell(0, 2, tview.NewTableCell("TAGS").SetSelectable(false))
	wl.Table.SetCell(0, 3, tview.NewTableCell("TERRAFORM").SetSelectable(false))
	wl.Table.SetCell(0, 4, tview.NewTableCell("COUNT").SetSelectable(false))
	wl.Table.SetCell(0, 5, tview.NewTableCell("RUN STATUS").SetSelectable(false))
	wl.Table.SetCell(0, 5, tview.NewTableCell("LATEST CHANGE").SetSelectable(false))
	for i, w := range wl.workspaces.Items {
		r := i + 1
		wl.Table.SetCell(r, 0, tview.NewTableCell(w.ID).SetExpansion(1))
		wl.Table.SetCell(r, 1, tview.NewTableCell(w.Name).SetExpansion(1))
		wl.Table.SetCell(r, 2, fmtTags(w).SetExpansion(1))
		wl.Table.SetCell(r, 3, tview.NewTableCell(w.TerraformVersion).SetExpansion(2))
		wl.Table.SetCell(r, 4, tview.NewTableCell(fmt.Sprint(w.ResourceCount)).SetExpansion(2))
		wl.Table.SetCell(r, 5, fmtCurrentRun(w).SetExpansion(1))
		wl.Table.SetCell(r, 6, fmtUpdatedAt(w).SetExpansion(1))
	}

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
	return KeyActions{}
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
