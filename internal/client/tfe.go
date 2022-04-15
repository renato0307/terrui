package client

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/hashicorp/go-tfe"
)

type TFEClient interface {
	ListOrganizations(pageNumber int) (*tfe.OrganizationList, error)
	ListWorkspaces(org string, searchText string, pageNumber int) (*tfe.WorkspaceList, error)
	ReadWorkspace(org, workspace string) (*tfe.Workspace, error)
	ListWorkspaceVariables(workspaceID string) (*tfe.VariableList, error)
	ListWorkspaceRuns(workspaceID string) (*tfe.RunList, error)
	ListWorkspaceTeamAccesses(workspaceID string) (*tfe.TeamAccessList, error)
	ReadWorkspaceRun(runID string) (*tfe.Run, error)
	ReadWorkspacePlan(planID string) (*tfe.Plan, error)
	ReadWorkspacePlanLogs(planID string) (io.Reader, error)
	ReadWorkspaceApplyLogs(planID string) (io.Reader, error)
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

	textSearch, tagsSearch := parseSearchText(searchText)

	return c.client.Workspaces.List(context.Background(), org, &tfe.WorkspaceListOptions{
		Include:     []tfe.WSIncludeOpt{"current_run"},
		Search:      textSearch,
		Tags:        tagsSearch,
		ListOptions: options,
	})
}

func parseSearchText(searchText string) (string, string) {
	tagsLabels := []string{"tags", "tag", "t"}
	b := strings.Builder{}
	var tagsSearch string
	for _, s := range strings.Split(searchText, " ") {
		var tags string
		var found bool

		for _, l := range tagsLabels {
			_, tags, found = strings.Cut(s, fmt.Sprintf("%s:", l))
			if found {
				break
			}
		}

		if found {
			tagsSearch = tags
		} else {
			b.WriteString(s)
			b.WriteString(" ")
		}
	}
	return strings.TrimRight(b.String(), " "), tagsSearch
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

func (c *TFEClientImpl) ListWorkspaceTeamAccesses(workspaceID string) (*tfe.TeamAccessList, error) {
	accesses, err := c.client.TeamAccess.List(context.Background(), &tfe.TeamAccessListOptions{
		WorkspaceID: workspaceID,
	})

	if err != nil {
		return nil, err
	}

	for _, a := range accesses.Items {
		team, err := c.client.Teams.Read(context.Background(), a.Team.ID)
		if err != nil {
			return nil, err
		}

		a.Team = team
	}

	return accesses, err
}

func (c *TFEClientImpl) ListWorkspaceVariables(workspaceID string) (*tfe.VariableList, error) {
	return c.client.Variables.List(context.Background(), workspaceID, &tfe.VariableListOptions{})
}

func (c *TFEClientImpl) ListWorkspaceRuns(workspaceID string) (*tfe.RunList, error) {
	options := &tfe.RunListOptions{Include: []tfe.RunIncludeOpt{"created_by"}}
	return c.client.Runs.List(context.Background(), workspaceID, options)
}

func (c *TFEClientImpl) ReadWorkspaceRun(runID string) (*tfe.Run, error) {
	options := &tfe.RunReadOptions{Include: []tfe.RunIncludeOpt{"plan", "apply"}}
	return c.client.Runs.ReadWithOptions(context.Background(), runID, options)
}

func (c *TFEClientImpl) ReadWorkspacePlan(planID string) (*tfe.Plan, error) {
	return c.client.Plans.Read(context.Background(), planID)
}

func (c *TFEClientImpl) ReadWorkspacePlanLogs(planID string) (io.Reader, error) {
	return c.client.Plans.Logs(context.Background(), planID)
}

func (c *TFEClientImpl) ReadWorkspaceApplyLogs(planID string) (io.Reader, error) {
	return c.client.Applies.Logs(context.Background(), planID)
}
