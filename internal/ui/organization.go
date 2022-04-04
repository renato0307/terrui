package ui

import (
	"fmt"

	"github.com/hashicorp/go-tfe"
	"github.com/rivo/tview"

	"github.com/renato0307/terrui/internal/client"
)

const OrganizationPageName string = "organization"

type OrganizationPage struct {
	*tview.Table

	app                 *App
	currentOrganization int
	orgs                *tfe.OrganizationList
}

func NewOrganizationPage(app *App) Page {
	ol := OrganizationPage{
		Table:               tview.NewTable(),
		app:                 app,
		currentOrganization: 1,
	}

	return &ol
}

func (o *OrganizationPage) Load() error {
	tfeClient, err := client.NewTFEClient()
	if err != nil {
		return fmt.Errorf("error creating the TFE client: %w", err)
	}

	orgs, err := tfeClient.ListOrganizations()
	if err != nil {
		return fmt.Errorf("error listing the organization: %w", err)
	}
	o.orgs = orgs

	return nil
}

func (o *OrganizationPage) View() string {
	o.SetSelectable(true, false)
	o.SetCell(0, 0, tview.NewTableCell("ID").SetSelectable(false))
	o.SetCell(0, 1, tview.NewTableCell("NAME").SetSelectable(false))
	o.SetCell(0, 2, tview.NewTableCell("E-MAIL").SetSelectable(false))
	for i, org := range o.orgs.Items {
		r := i + 1
		o.SetCell(r, 0, tview.NewTableCell(org.ExternalID).SetExpansion(1))
		o.SetCell(r, 1, tview.NewTableCell(org.Name).SetExpansion(1))
		o.SetCell(r, 2, tview.NewTableCell(org.Email).SetExpansion(2))
	}

	o.SetSelectionChangedFunc(func(row, column int) {
		o.currentOrganization = row + 1
	})

	return "organization loaded"
}

func (o *OrganizationPage) BindKeys() KeyActions {
	return KeyActions{}
}
