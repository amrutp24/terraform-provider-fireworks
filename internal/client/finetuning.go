package client

import (
	"context"
	"fmt"
	"net/http"
)

// Fine-tuning job states reported by the API.
const (
	JobStateCreating  = "JOB_STATE_CREATING"
	JobStateRunning   = "JOB_STATE_RUNNING"
	JobStateCompleted = "JOB_STATE_COMPLETED"
	JobStateFailed    = "JOB_STATE_FAILED"
	JobStateCancelled = "JOB_STATE_CANCELLED"
)

// SupervisedFineTuningJob mirrors the API object. Jobs are immutable once
// created. Pointers distinguish "unset" from zero values.
type SupervisedFineTuningJob struct {
	Name              string   `json:"name,omitempty"`
	DisplayName       string   `json:"displayName,omitempty"`
	Dataset           string   `json:"dataset,omitempty"`
	BaseModel         string   `json:"baseModel,omitempty"`
	WarmStartFrom     string   `json:"warmStartFrom,omitempty"`
	OutputModel       string   `json:"outputModel,omitempty"`
	EvaluationDataset string   `json:"evaluationDataset,omitempty"`
	Epochs            *int64   `json:"epochs,omitempty"`
	LearningRate      *float64 `json:"learningRate,omitempty"`
	LoraRank          *int64   `json:"loraRank,omitempty"`
	BatchSize         *int64   `json:"batchSize,omitempty"`
	MaxContextLength  *int64   `json:"maxContextLength,omitempty"`
	EarlyStop         bool     `json:"earlyStop,omitempty"`

	// Read-only.
	State         string `json:"state,omitempty"`
	CreateTime    string `json:"createTime,omitempty"`
	CompletedTime string `json:"completedTime,omitempty"`
	CreatedBy     string `json:"createdBy,omitempty"`
}

func (c *Client) sftJobsPath() string {
	return fmt.Sprintf("/v1/accounts/%s/supervisedFineTuningJobs", c.accountID)
}

func (c *Client) sftJobPath(id string) string {
	return c.sftJobsPath() + "/" + id
}

func (c *Client) CreateSupervisedFineTuningJob(ctx context.Context, job *SupervisedFineTuningJob) (*SupervisedFineTuningJob, error) {
	var out SupervisedFineTuningJob
	if err := c.do(ctx, http.MethodPost, c.sftJobsPath(), job, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetSupervisedFineTuningJob(ctx context.Context, id string) (*SupervisedFineTuningJob, error) {
	var out SupervisedFineTuningJob
	if err := c.do(ctx, http.MethodGet, c.sftJobPath(id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteSupervisedFineTuningJob(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, c.sftJobPath(id), nil, nil)
}
