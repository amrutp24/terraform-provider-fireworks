---
page_title: "fireworks_model Data Source"
description: |-
  Look up a model in the configured account by ID.
---

# fireworks_model (Data Source)

Look up a model in the configured account by ID — typically the output model
of a fine-tuning job or an uploaded custom model.

## Example Usage

```terraform
data "fireworks_model" "tuned" {
  id = "my-fine-tuned-model"
}

resource "fireworks_deployment" "tuned" {
  base_model = data.fireworks_model.tuned.name
}
```

## Schema

### Required

- `id` (String) Model ID (last segment of the resource name).

### Read-Only

- `name` (String) Full resource name.
- `display_name` (String) Human-readable name.
- `description` (String) Model description.
- `state` (String) Model state (`UPLOADING`, `READY`).
- `kind` (String) Model kind, e.g. `HF_BASE_MODEL`, `HF_PEFT_ADDON`.
- `create_time` (String) Creation timestamp.
