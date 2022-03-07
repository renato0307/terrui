package ui

import (
	"context"
	"log"
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/hashicorp/go-tfe"
	"github.com/rivo/tview"
)

type OrganizationList struct {
	*tview.Table
}

func NewOrganizationList() *OrganizationList {
	ol := OrganizationList{
		Table: tview.NewTable(),
	}

	loading := tview.NewTableCell("loading...").
		SetAlign(tview.AlignCenter).
		SetTextColor(tcell.ColorPaleVioletRed)
	ol.Table.SetCell(0, 0, loading.SetExpansion(1))

	return &ol
}

func (ol *OrganizationList) Execute() {
	config := &tfe.Config{
		Token: os.Getenv("TFE_TOKEN"),
	}

	client, err := tfe.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}

	orgs, err := client.Organizations.List(context.Background(), tfe.OrganizationListOptions{})
	if err != nil {
		log.Fatal(err)
	}

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

	ol.SetInputCapture(ol.keyboard)
}

func (ol *OrganizationList) keyboard(evt *tcell.EventKey) *tcell.EventKey {
	return evt
}
