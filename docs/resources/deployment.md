---
page_title: "fireworks_deployment Resource"
description: |-
  An on-demand GPU inference deployment.
---

# fireworks_deployment (Resource)

An on-demand GPU inference deployment. Billed per GPU-minute while replicas
are running; set `min_replica_count = 0` to allow scale-to-zero.

## Example Usage

```terraform
resource "fireworks_deployment" "llama" {
  base_model        = "accounts/fireworks/models/llama-v3p1-8b-instruct"
  display_name      = "prod inference"
  min_replica_count = 0
  max_replica_count = 2
  accelerator_type  = "NVIDIA_H100_80GB"
  precision         = "FP8"
  region            = "US_IOWA_1"
}
```

## Schema

### Required

- `base_model` (String) Model to deploy, e.g. `accounts/fireworks/models/llama-v3p1-8b-instruct`.

### Optional

- `display_name` (String) Human-readable name (max 64 characters).
- `description` (String) Description of the deployment.
- `min_replica_count` (Number) Minimum replicas. `0` enables scale-to-zero. API default: `0`.
- `max_replica_count` (Number) Maximum replicas for autoscaling. API default: `max(min_replica_count, 1)`.
- `accelerator_type` (String) GPU type, e.g. `NVIDIA_A100_80GB`, `NVIDIA_H100_80GB`, `NVIDIA_H200_141GB`, `AMD_MI300X_192GB`.
- `accelerator_count` (Number) GPUs per replica.
- `precision` (String) Model serving precision, e.g. `FP16`, `FP8`, `BF16`, `NF4`.
- `region` (String) Placement region, e.g. `US_IOWA_1`, `EU_FRANKFURT_1`.
- `wait_for_ready` (Boolean) Whether to wait until the deployment reaches `READY`. Defaults to `true`.

### Read-Only

- `id` (String) Deployment ID.
- `name` (String) Full resource name, e.g. `accounts/my-account/deployments/abc123`.
- `state` (String) Current deployment state.
- `create_time` (String) Creation timestamp (RFC 3339).

## Import

```shell
terraform import fireworks_deployment.llama <deployment-id>
```
