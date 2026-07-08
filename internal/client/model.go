package client

import (
	"context"
	"fmt"
	"net/http"
)

// Model mirrors the read-side of the gatewayModel API object.
type Model struct {
	Name        string `json:"name,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
	Description string `json:"description,omitempty"`
	State       string `json:"state,omitempty"`
	Kind        string `json:"kind,omitempty"`
	CreateTime  string `json:"createTime,omitempty"`
}

func (c *Client) GetModel(ctx context.Context, id string) (*Model, error) {
	var out Model
	path := fmt.Sprintf("/v1/accounts/%s/models/%s", c.accountID, id)
	if err := c.do(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
