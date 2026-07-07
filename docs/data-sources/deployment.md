---
page_title: "fireworks_deployment Data Source"
description: |-
  Look up an existing deployment by ID.
---

# fireworks_deployment (Data Source)

Look up an existing deployment by ID.

## Example Usage

```terraform
data "fireworks_deployment" "existing" {
  id = "abc123"
}

output "model" {
  value = data.fireworks_deployment.existing.base_model
}
```

## Schema

### Required

- `id` (String) Deployment ID.

### Read-Only

- `name` (String) Full resource name.
- `base_model` (String) Deployed model.
- `display_name` (String) Human-readable name.
- `min_replica_count` (Number) Minimum replicas.
- `max_replica_count` (Number) Maximum replicas.
- `accelerator_type` (String) GPU type.
- `precision` (String) Serving precision.
- `state` (String) Current state.
