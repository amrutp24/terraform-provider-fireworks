---
page_title: "fireworks_dataset Resource"
description: |-
  A dataset for fine-tuning or evaluation.
---

# fireworks_dataset (Resource)

A dataset for fine-tuning or evaluation. Attributes are immutable; changes
force replacement.

## Example Usage

```terraform
resource "fireworks_dataset" "training" {
  dataset_id    = "training-data"
  display_name  = "Training data"
  external_url  = "gs://my-bucket/train.jsonl"
  format        = "CHAT"
  example_count = 1200
}
```

## Schema

### Required

- `dataset_id` (String) User-chosen dataset identifier. Forces replacement.

### Optional

- `display_name` (String) Human-readable name. Forces replacement.
- `external_url` (String) External URI of the dataset contents, e.g. `gs://bucket/data.jsonl`. Forces replacement.
- `format` (String) Dataset format: `CHAT`, `COMPLETION`, or `RL`. Forces replacement.
- `example_count` (Number) Number of examples in the dataset. The API requires
  it when `external_url` is not set (uploaded datasets). Forces replacement.

### Read-Only

- `id` (String) Same as `dataset_id`.
- `name` (String) Full resource name.
- `state` (String) Dataset state (`UPLOADING`, `READY`).

## Import

```shell
terraform import fireworks_dataset.training <dataset-id>
```
