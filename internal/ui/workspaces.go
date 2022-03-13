package ui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/hashicorp/go-tfe"
	"github.com/rivo/tview"

	"github.com/renato0307/terrui/internal/client"
)

type WorkspaceList struct {
	*tview.Table

	app          *App
	tfeClient    *client.TFEClient
	organization string
}

func NewWorkspaceList(app *App, organization string) (*WorkspaceList, error) {
	tfeClient, err := client.NewTFEClient()
	if err != nil {
		app.message.ShowError("could not show organizations")
		return nil, fmt.Errorf("error creating the TFE client: %w", err)
	}

	ol := WorkspaceList{
		Table: tview.NewTable(),

		app:          app,
		tfeClient:    tfeClient,
		organization: organization,
	}

	return &ol, nil
}

func (wl *WorkspaceList) Load() {

	wl.app.QueueUpdateDraw(func() {
		wl.app.message.ShowText("> workspaces")
		loading := tview.NewTableCell("loading workspaces...").
			SetAlign(tview.AlignCenter).
			SetTextColor(tcell.ColorPaleVioletRed)
		wl.Table.SetCell(0, 0, loading.SetExpansion(1))

		wl.app.supportedCmds.SetCommands(
			[]SupportedCommand{
				// {
				// 	ShortCut: "d",
				// 	Name:     "show details",
				// },
				{
					ShortCut: "enter",
					Name:     "select workspace",
				},
			},
		)
	})

	orgs, err := wl.tfeClient.ListWorkspaces(wl.organization)
	if err != nil {
		wl.app.message.ShowError("could not fetch workspaces")

		wl.app.QueueUpdateDraw(func() {
			loading := tview.NewTableCell("😵 error loading workspaces").
				SetAlign(tview.AlignCenter).
				SetTextColor(tcell.ColorPaleVioletRed)
			wl.Table.SetCell(0, 0, loading.SetExpansion(1))
		})
		return
	}

	wl.app.QueueUpdateDraw(func() {
		wl.Table.SetSelectable(true, false)
		wl.Table.SetCell(0, 0, tview.NewTableCell("NAME").SetSelectable(false))
		wl.Table.SetCell(0, 1, tview.NewTableCell("TAGS").SetSelectable(false))
		wl.Table.SetCell(0, 2, tview.NewTableCell("TERRAFORM").SetSelectable(false))
		wl.Table.SetCell(0, 3, tview.NewTableCell("COUNT").SetSelectable(false))
		wl.Table.SetCell(0, 4, tview.NewTableCell("RUN STATUS").SetSelectable(false))
		wl.Table.SetCell(0, 5, tview.NewTableCell("LATEST CHANGE").SetSelectable(false))
		for i, w := range orgs.Items {
			r := i + 1
			wl.Table.SetCell(r, 0, tview.NewTableCell(w.Name).SetExpansion(1))
			wl.Table.SetCell(r, 1, fmtTags(w).SetExpansion(1))
			wl.Table.SetCell(r, 2, tview.NewTableCell(w.TerraformVersion).SetExpansion(2))
			wl.Table.SetCell(r, 3, tview.NewTableCell(fmt.Sprint(w.ResourceCount)).SetExpansion(2))
			wl.Table.SetCell(r, 4, fmtCurrentRun(w).SetExpansion(1))
			wl.Table.SetCell(r, 5, fmtUpdatedAt(w).SetExpansion(1))
		}
	})

	wl.app.SetFocus(wl)
	wl.SetInputCapture(wl.keyboard)
}

func (wl *WorkspaceList) keyboard(evt *tcell.EventKey) *tcell.EventKey {
	// nolint:exhaustive
	switch evt.Key() {
	case tcell.KeyEnter, tcell.KeyCtrlE:
	}

	return evt
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

func fmtTags(w *tfe.Workspace) *tview.TableCell {
	s := fmt.Sprint(w.TagNames)
	return tview.NewTableCell(strings.Trim(s, "[]"))
}

func fmtUpdatedAt(w *tfe.Workspace) *tview.TableCell {
	return tview.NewTableCell(w.UpdatedAt.Local().Format("02 Jan 06 15:04 MST"))
}
