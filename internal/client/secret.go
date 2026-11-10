package client

import (
	"context"
	"fmt"
	"net/http"
)

// Secret mirrors the gatewaySecret API object. Value is input-only: the API
// accepts it on create and update but never returns it.
type Secret struct {
	Name    string `json:"name,omitempty"`
	KeyName string `json:"keyName,omitempty"`
	Value   string `json:"value,omitempty"`
}

func (c *Client) secretsPath() string {
	return fmt.Sprintf("/v1/accounts/%s/secrets", c.accountID)
}

func (c *Client) secretPath(id string) string {
	return c.secretsPath() + "/" + id
}

// SecretName returns the full resource name for a secret ID in this account.
func (c *Client) SecretName(id string) string {
	return fmt.Sprintf("accounts/%s/secrets/%s", c.accountID, id)
}

func (c *Client) CreateSecret(ctx context.Context, id, keyName, value string) (*Secret, error) {
	var out Secret
	in := &Secret{Name: c.SecretName(id), KeyName: keyName, Value: value}
	if err := c.do(ctx, http.MethodPost, c.secretsPath(), in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetSecret(ctx context.Context, id string) (*Secret, error) {
	var out Secret
	if err := c.do(ctx, http.MethodGet, c.secretPath(id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UpdateSecret(ctx context.Context, id, keyName, value string) (*Secret, error) {
	var out Secret
	in := &Secret{KeyName: keyName, Value: value}
	if err := c.do(ctx, http.MethodPatch, c.secretPath(id), in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteSecret(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, c.secretPath(id), nil, nil)
}
