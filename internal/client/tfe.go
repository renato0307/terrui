package client

import (
	"context"
	"os"

	"github.com/hashicorp/go-tfe"
)

type TFEClient interface {
	ListOrganizations(pageNumber int) (*tfe.OrganizationList, error)
	ListWorkspaces(org string, searchText string, pageNumber int) (*tfe.WorkspaceList, error)
	ReadWorkspace(org, workspace string) (*tfe.Workspace, error)
	ListWorkspaceVariables(workspaceID string) (*tfe.VariableList, error)
}

type TFEClientImpl struct {
	config *tfe.Config
	client *tfe.Client
}

func NewTFEClient() (TFEClient, error) {
	c := TFEClientImpl{}

	c.config = &tfe.Config{Token: os.Getenv("TFE_TOKEN")}

	client, err := tfe.NewClient(c.config)
	if err != nil {
		return nil, err
	}
	c.client = client

	return &c, nil
}

func (c *TFEClientImpl) ListOrganizations(pageNumber int) (*tfe.OrganizationList, error) {
	options := tfe.OrganizationListOptions{
		ListOptions: tfe.ListOptions{
			PageSize: 30,
		},
	}
	if pageNumber != -1 {
		options.PageNumber = pageNumber
	}

	return c.client.Organizations.List(context.Background(), &options)
}

func (c *TFEClientImpl) ListWorkspaces(org string, searchText string, pageNumber int) (*tfe.WorkspaceList, error) {
	options := tfe.ListOptions{PageSize: 30}
	if pageNumber != -1 {
		options.PageNumber = pageNumber
	}

	return c.client.Workspaces.List(context.Background(), org, &tfe.WorkspaceListOptions{
		Include:     []tfe.WSIncludeOpt{"current_run"},
		Search:      searchText,
		ListOptions: options,
	})
}

func (c *TFEClientImpl) ReadWorkspace(org, workspace string) (*tfe.Workspace, error) {
	w, err := c.client.Workspaces.ReadWithOptions(context.Background(), org, workspace, &tfe.WorkspaceReadOptions{
		Include: []tfe.WSIncludeOpt{"current_run", "current_run.plan", "locked_by"},
	})
	if err != nil {
		return nil, err
	}

	r, err := c.client.Runs.ReadWithOptions(context.Background(), w.CurrentRun.ID, &tfe.RunReadOptions{
		Include: []tfe.RunIncludeOpt{"created_by", "plan", "apply"},
	})
	if err != nil {
		return nil, err
	}
	w.CurrentRun = r

	return w, err
}

func (c *TFEClientImpl) ListWorkspaceVariables(workspaceID string) (*tfe.VariableList, error) {
	return c.client.Variables.List(context.Background(), workspaceID, &tfe.VariableListOptions{})
}
