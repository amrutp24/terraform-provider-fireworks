package client

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// Deployed model states reported by the API. There is no failed state; load
// errors are reported in Status.
const (
	DeployedModelStateDeploying   = "DEPLOYING"
	DeployedModelStateDeployed    = "DEPLOYED"
	DeployedModelStateUpdating    = "UPDATING"
	DeployedModelStateUndeploying = "UNDEPLOYING"
)

// DeployedModel mirrors the gatewayDeployedModel API object: a LoRA addon
// loaded onto a dedicated deployment or onto serverless. Pointer bools
// distinguish "unset" from false.
type DeployedModel struct {
	Name        string `json:"name,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
	Description string `json:"description,omitempty"`
	Model       string `json:"model,omitempty"`
	Deployment  string `json:"deployment,omitempty"`
	Default     *bool  `json:"default,omitempty"`
	Serverless  *bool  `json:"serverless,omitempty"`
	Public      *bool  `json:"public,omitempty"`

	// Read-only.
	State      string  `json:"state,omitempty"`
	Status     *Status `json:"status,omitempty"`
	CreateTime string  `json:"createTime,omitempty"`
	UpdateTime string  `json:"updateTime,omitempty"`
}

// Status mirrors gatewayStatus (google.rpc.Status).
type Status struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

// DeployedModelUpdate is the PATCH body. displayName and description are
// always sent so they can be cleared.
type DeployedModelUpdate struct {
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
	Default     *bool  `json:"default,omitempty"`
	Public      *bool  `json:"public,omitempty"`
}

func (c *Client) deployedModelsPath() string {
	return fmt.Sprintf("/v1/accounts/%s/deployedModels", c.accountID)
}

func (c *Client) deployedModelPath(id string) string {
	return c.deployedModelsPath() + "/" + id
}

// CreateDeployedModel loads a LoRA addon. replaceMergedAddon sets the query
// parameter of the same name, which hot-reload deployments require.
func (c *Client) CreateDeployedModel(ctx context.Context, dm *DeployedModel, replaceMergedAddon bool) (*DeployedModel, error) {
	p := c.deployedModelsPath()
	if replaceMergedAddon {
		p += "?replaceMergedAddon=true"
	}
	var out DeployedModel
	if err := c.do(ctx, http.MethodPost, p, dm, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetDeployedModel(ctx context.Context, id string) (*DeployedModel, error) {
	var out DeployedModel
	if err := c.do(ctx, http.MethodGet, c.deployedModelPath(id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UpdateDeployedModel(ctx context.Context, id string, upd *DeployedModelUpdate) (*DeployedModel, error) {
	var out DeployedModel
	if err := c.do(ctx, http.MethodPatch, c.deployedModelPath(id), upd, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteDeployedModel unloads the LoRA addon.
func (c *Client) DeleteDeployedModel(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, c.deployedModelPath(id), nil, nil)
}

// WaitForDeployedModelReady polls until the deployed model reaches DEPLOYED,
// returning the final object.
func (c *Client) WaitForDeployedModelReady(ctx context.Context, id string) (*DeployedModel, error) {
	for {
		dm, err := c.GetDeployedModel(ctx, id)
		if err != nil {
			return nil, err
		}
		if dm.State == DeployedModelStateDeployed {
			return dm, nil
		}

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timed out waiting for deployed model %s to become DEPLOYED (last state: %s%s)",
				id, dm.State, statusSuffix(dm.Status))
		case <-time.After(5 * time.Second):
		}
	}
}

// WaitForDeployedModelGone polls until the deployed model no longer exists.
func (c *Client) WaitForDeployedModelGone(ctx context.Context, id string) error {
	for {
		dm, err := c.GetDeployedModel(ctx, id)
		if err != nil {
			if IsNotFound(err) {
				return nil
			}
			return err
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("timed out waiting for deployed model %s to be unloaded (last state: %s%s)",
				id, dm.State, statusSuffix(dm.Status))
		case <-time.After(5 * time.Second):
		}
	}
}

func statusSuffix(s *Status) string {
	if s == nil || s.Message == "" {
		return ""
	}
	return fmt.Sprintf(", status: %s %s", s.Code, s.Message)
}
