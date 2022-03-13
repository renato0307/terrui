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
