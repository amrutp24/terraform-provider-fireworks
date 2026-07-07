---
page_title: "Fireworks Provider"
description: |-
  Manage Fireworks AI resources: GPU inference deployments and datasets.
---

# Fireworks Provider

Manage [Fireworks AI](https://fireworks.ai) control-plane resources with
Terraform: on-demand GPU inference deployments and datasets.

## Example Usage

```terraform
terraform {
  required_providers {
    fireworks = {
      source = "amrutp24/fireworks"
    }
  }
}

provider "fireworks" {
  # api_key and account_id can also come from FIREWORKS_API_KEY
  # and FIREWORKS_ACCOUNT_ID environment variables.
  account_id = "my-account"
}

resource "fireworks_deployment" "llama" {
  base_model        = "accounts/fireworks/models/llama-v3p1-8b-instruct"
  display_name      = "prod inference"
  min_replica_count = 0 # scale to zero when idle
  max_replica_count = 2
  accelerator_type  = "NVIDIA_H100_80GB"
}
```

## Schema

### Optional

- `api_key` (String, Sensitive) Fireworks API key. Falls back to the
  `FIREWORKS_API_KEY` environment variable.
- `account_id` (String) Fireworks account ID. Falls back to the
  `FIREWORKS_ACCOUNT_ID` environment variable.
- `base_url` (String) Override the API base URL. Defaults to
  `https://api.fireworks.ai`. Mainly for testing.
