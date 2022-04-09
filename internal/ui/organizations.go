package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/hashicorp/go-tfe"
	"github.com/rivo/tview"

	"github.com/renato0307/terrui/internal/client"
)

const OrganizationsPageName string = "organizations"

type OrganizationsPageSource struct {
	app  *App
	orgs *tfe.OrganizationList
}

func NewOrganizationsPage(app *App) Page {
	return NewListPage(app, &OrganizationsPageSource{app: app})
}

func (o *OrganizationsPageSource) SupportsSearch() bool {
	return false
}

func (o *OrganizationsPageSource) Search(searchText string, pageNumber int) error {
	tfeClient, err := client.NewTFEClient()
	if err != nil {
		return fmt.Errorf("error creating the TFE client: %w", err)
	}

	orgs, err := tfeClient.ListOrganizations(pageNumber)
	if err != nil {
		return fmt.Errorf("error listing the organization: %w", err)
	}
	o.orgs = orgs

	return nil
}

func (o *OrganizationsPageSource) RenderHeader(table *tview.Table) {
	table.SetCell(0, 0, tview.NewTableCell("ID").SetSelectable(false))
	table.SetCell(0, 1, tview.NewTableCell("NAME").SetSelectable(false))
	table.SetCell(0, 2, tview.NewTableCell("E-MAIL").SetSelectable(false))
}

func (o *OrganizationsPageSource) RenderRows(table *tview.Table) {
	for i, org := range o.orgs.Items {
		r := i + 1
		table.SetCell(r, 0, tview.NewTableCell(org.ExternalID).SetExpansion(1))
		table.SetCell(r, 1, tview.NewTableCell(org.Name).SetExpansion(1))
		table.SetCell(r, 2, tview.NewTableCell(org.Email).SetExpansion(2))
	}
}

func (o *OrganizationsPageSource) Crumb() []string {
	return []string{OrganizationsPageName}
}

func (o *OrganizationsPageSource) ActionSelectWorkspace(table *tview.Table, currentItem int) func(ek *tcell.EventKey) *tcell.EventKey {
	return func(ek *tcell.EventKey) *tcell.EventKey {
		o.app.config.Organization = table.GetCell(currentItem, 1).Text
		o.app.config.Save()

		o.app.activatePage(WorkspacesPageName, nil, false)

		return nil
	}
}

func (o *OrganizationsPageSource) Name() string {
	return "organization"
}

func (o *OrganizationsPageSource) NameList() string {
	return OrganizationsPageName
}

func (o *OrganizationsPageSource) Empty() bool {
	return o.orgs == nil || len(o.orgs.Items) == 0
}

func (o *OrganizationsPageSource) CurrentPage() int {
	return o.orgs.CurrentPage
}

func (o *OrganizationsPageSource) TotalCount() int {
	return o.orgs.TotalCount
}

func (o *OrganizationsPageSource) TotalPages() int {
	return o.orgs.TotalPages
}
