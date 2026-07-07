package client

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// Deployment states reported by the API.
const (
	DeploymentStateCreating = "CREATING"
	DeploymentStateReady    = "READY"
	DeploymentStateUpdating = "UPDATING"
	DeploymentStateDeleting = "DELETING"
	DeploymentStateFailed   = "FAILED"
	DeploymentStateDeleted  = "DELETED"
)

// Deployment mirrors the gatewayDeployment API object. Pointers distinguish
// "unset" from zero values (e.g. minReplicaCount: 0 is meaningful).
type Deployment struct {
	Name             string     `json:"name,omitempty"`
	DisplayName      string     `json:"displayName,omitempty"`
	Description      string     `json:"description,omitempty"`
	BaseModel        string     `json:"baseModel,omitempty"`
	MinReplicaCount  *int64     `json:"minReplicaCount,omitempty"`
	MaxReplicaCount  *int64     `json:"maxReplicaCount,omitempty"`
	AcceleratorType  string     `json:"acceleratorType,omitempty"`
	AcceleratorCount *int64     `json:"acceleratorCount,omitempty"`
	Precision        string     `json:"precision,omitempty"`
	Placement        *Placement `json:"placement,omitempty"`

	// Read-only.
	State      string `json:"state,omitempty"`
	CreateTime string `json:"createTime,omitempty"`
}

type Placement struct {
	Region string `json:"region,omitempty"`
}

func (c *Client) deploymentsPath() string {
	return fmt.Sprintf("/v1/accounts/%s/deployments", c.accountID)
}

func (c *Client) deploymentPath(id string) string {
	return c.deploymentsPath() + "/" + id
}

func (c *Client) CreateDeployment(ctx context.Context, dep *Deployment) (*Deployment, error) {
	var out Deployment
	if err := c.do(ctx, http.MethodPost, c.deploymentsPath(), dep, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetDeployment(ctx context.Context, id string) (*Deployment, error) {
	var out Deployment
	if err := c.do(ctx, http.MethodGet, c.deploymentPath(id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UpdateDeployment(ctx context.Context, id string, dep *Deployment) (*Deployment, error) {
	var out Deployment
	if err := c.do(ctx, http.MethodPatch, c.deploymentPath(id), dep, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteDeployment(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, c.deploymentPath(id), nil, nil)
}

// WaitForDeploymentReady polls until the deployment reaches READY, returning
// the final object. FAILED is terminal and returns an error.
func (c *Client) WaitForDeploymentReady(ctx context.Context, id string) (*Deployment, error) {
	for {
		dep, err := c.GetDeployment(ctx, id)
		if err != nil {
			return nil, err
		}
		switch dep.State {
		case DeploymentStateReady:
			return dep, nil
		case DeploymentStateFailed:
			return nil, fmt.Errorf("deployment %s entered FAILED state", id)
		}

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timed out waiting for deployment %s to become READY (last state: %s)", id, dep.State)
		case <-time.After(5 * time.Second):
		}
	}
}

// WaitForDeploymentGone polls until the deployment is deleted.
func (c *Client) WaitForDeploymentGone(ctx context.Context, id string) error {
	for {
		dep, err := c.GetDeployment(ctx, id)
		if err != nil {
			if IsNotFound(err) {
				return nil
			}
			return err
		}
		if dep.State == DeploymentStateDeleted {
			return nil
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("timed out waiting for deployment %s to be deleted (last state: %s)", id, dep.State)
		case <-time.After(5 * time.Second):
		}
	}
}
