package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/renato0307/terrui/internal/client"
)

type OrganizationList struct {
	*tview.Table

	app       *App
	tfeClient *client.TFEClient
}

func NewOrganizationList(app *App) (*OrganizationList, error) {
	tfeClient, err := client.NewTFEClient()
	if err != nil {
		app.message.ShowError("could not show organizations")
		return nil, fmt.Errorf("error creating the TFE client: %w", err)
	}

	ol := OrganizationList{
		Table: tview.NewTable(),

		app:       app,
		tfeClient: tfeClient,
	}

	return &ol, nil
}

func (ol *OrganizationList) Load() {

	ol.app.QueueUpdateDraw(func() {
		loading := tview.NewTableCell("loading...").
			SetAlign(tview.AlignCenter).
			SetTextColor(tcell.ColorPaleVioletRed)
		ol.Table.SetCell(0, 0, loading.SetExpansion(1))
	})

	orgs, err := ol.tfeClient.ListOrganizations()
	if err != nil {
		ol.app.message.ShowError("could not fetch organizations")

		ol.app.QueueUpdateDraw(func() {
			loading := tview.NewTableCell("😵 error loading organizations").
				SetAlign(tview.AlignCenter).
				SetTextColor(tcell.ColorPaleVioletRed)
			ol.Table.SetCell(0, 0, loading.SetExpansion(1))
		})
		return
	}

	ol.app.QueueUpdateDraw(func() {
		ol.Table.SetSelectable(true, false)
		ol.Table.SetCell(0, 0, tview.NewTableCell("ID").SetSelectable(false))
		ol.Table.SetCell(0, 1, tview.NewTableCell("NAME").SetSelectable(false))
		ol.Table.SetCell(0, 2, tview.NewTableCell("E-MAIL").SetSelectable(false))
		for i, o := range orgs.Items {
			r := i + 1
			ol.Table.SetCell(r, 0, tview.NewTableCell(o.ExternalID).SetExpansion(1))
			ol.Table.SetCell(r, 1, tview.NewTableCell(o.Name).SetExpansion(1))
			ol.Table.SetCell(r, 2, tview.NewTableCell(o.Email).SetExpansion(2))
		}
	})

	ol.app.SetFocus(ol)
	ol.SetInputCapture(ol.keyboard)
}

func (ol *OrganizationList) keyboard(evt *tcell.EventKey) *tcell.EventKey {
	return evt
}
