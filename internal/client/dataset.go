package client

import (
	"context"
	"fmt"
	"net/http"
)

// Dataset mirrors the gatewayDataset API object.
type Dataset struct {
	Name         string `json:"name,omitempty"`
	DisplayName  string `json:"displayName,omitempty"`
	ExternalURL  string `json:"externalUrl,omitempty"`
	Format       string `json:"format,omitempty"`
	ExampleCount int64  `json:"exampleCount,omitempty,string"`

	// Read-only.
	State      string `json:"state,omitempty"`
	CreateTime string `json:"createTime,omitempty"`
}

type createDatasetRequest struct {
	Dataset   *Dataset `json:"dataset"`
	DatasetID string   `json:"datasetId"`
}

func (c *Client) datasetsPath() string {
	return fmt.Sprintf("/v1/accounts/%s/datasets", c.accountID)
}

func (c *Client) datasetPath(id string) string {
	return c.datasetsPath() + "/" + id
}

func (c *Client) CreateDataset(ctx context.Context, id string, ds *Dataset) (*Dataset, error) {
	var out Dataset
	req := createDatasetRequest{Dataset: ds, DatasetID: id}
	if err := c.do(ctx, http.MethodPost, c.datasetsPath(), req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetDataset(ctx context.Context, id string) (*Dataset, error) {
	var out Dataset
	if err := c.do(ctx, http.MethodGet, c.datasetPath(id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteDataset(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, c.datasetPath(id), nil, nil)
}
