package client

import (
	"context"
	"os"

	"github.com/hashicorp/go-tfe"
)

type TFEClient struct {
	config *tfe.Config
	client *tfe.Client
}

func NewTFEClient() (*TFEClient, error) {
	c := TFEClient{}

	c.config = &tfe.Config{Token: os.Getenv("TFE_TOKEN")}

	client, err := tfe.NewClient(c.config)
	if err != nil {
		return nil, err
	}
	c.client = client

	return &c, nil
}

func (c *TFEClient) ListOrganizations() (*tfe.OrganizationList, error) {
	return c.client.Organizations.List(context.Background(), tfe.OrganizationListOptions{})
}

func (c *TFEClient) ListWorkspaces(org string) (*tfe.WorkspaceList, error) {
	includes := "current_run"
	return c.client.Workspaces.List(context.Background(), org, tfe.WorkspaceListOptions{
		Include:     &includes,
		ListOptions: tfe.ListOptions{PageSize: 30},
	})
}

func (c *TFEClient) ReadWorkspace(org, workspace string) (*tfe.Workspace, error) {
	w, err := c.client.Workspaces.ReadWithOptions(context.Background(), org, workspace, &tfe.WorkspaceReadOptions{
		Include: "current_run,current_run.plan,locked_by",
	})
	if err != nil {
		return nil, err
	}

	r, err := c.client.Runs.ReadWithOptions(context.Background(), w.CurrentRun.ID, &tfe.RunReadOptions{
		Include: "created_by,plan,apply",
	})
	if err != nil {
		return nil, err
	}
	w.CurrentRun = r

	return w, err
}

func (c *TFEClient) ListWorkspaceVariables(workspaceID string) (*tfe.VariableList, error) {
	return c.client.Variables.List(context.Background(), workspaceID, tfe.VariableListOptions{})
}
