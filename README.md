# terraform-provider-fireworks

[![CI](https://github.com/amrutp24/terraform-provider-fireworks/actions/workflows/ci.yml/badge.svg)](https://github.com/amrutp24/terraform-provider-fireworks/actions/workflows/ci.yml)
[![Registry](https://img.shields.io/badge/registry-amrutp24%2Ffireworks-5c4ee5)](https://registry.terraform.io/providers/amrutp24/fireworks/latest)

Terraform provider for [Fireworks AI](https://fireworks.ai): GPU inference
deployments, fine-tuning datasets and fine-tuning jobs.

Docs: [registry.terraform.io/providers/amrutp24/fireworks/latest/docs](https://registry.terraform.io/providers/amrutp24/fireworks/latest/docs).
Changelog: [CHANGELOG.md](CHANGELOG.md).

## Maturity

A community provider at 0.x. Minor releases can change the schema, so pin a
version.

What has been run against the real API:

| Resource | Production |
| --- | --- |
| `fireworks_dataset` | create, refresh, destroy |
| `fireworks_deployment` | mock tests only; creating one needs a payment method on the account |
| `fireworks_supervised_fine_tuning_job` | mock tests only; training a real job costs money |
| `fireworks_model` | mock tests only |

Where the real API has not been exercised, the schema follows Fireworks'
documentation.

## Usage

```terraform
terraform {
  required_providers {
    fireworks = {
      source  = "amrutp24/fireworks"
      version = "~> 0.3"
    }
  }
}

provider "fireworks" {
  # Read from FIREWORKS_API_KEY and FIREWORKS_ACCOUNT_ID when unset here.
  account_id = "my-account"
}

# Datasets are immutable; any change replaces the resource.
resource "fireworks_dataset" "training" {
  dataset_id   = "support-tickets-v1"
  format       = "CHAT"
  external_url = "gs://my-bucket/support-tickets.jsonl"
}
```

### Deployments

A deployment reserves GPUs while replicas are up. `min_replica_count`
defaults to 0, which scales to zero when idle; anything above 0 is always-on
spend.

```terraform
resource "fireworks_deployment" "llama" {
  base_model       = "accounts/fireworks/models/llama-v3p1-8b-instruct"
  accelerator_type = "NVIDIA_H100_80GB"

  # Scale to zero; the first request afterwards pays a cold start.
  min_replica_count = 0
  max_replica_count = 2
}
```

`apply` waits for `READY`. Set `wait_for_ready = false` to return once the API
accepts the request.

GPU rates are on the [pricing page](https://fireworks.ai/pricing). Neither
this provider nor `terraform plan` reports cost.

## Resources

| Name | Notes |
| --- | --- |
| `fireworks_deployment` | On-demand GPU inference. CRUD, import, waits for `READY`. |
| `fireworks_dataset` | Fine-tuning or evaluation dataset. Immutable. Registers data by `external_url`; local file upload is not supported yet. |
| `fireworks_supervised_fine_tuning_job` | Asynchronous SFT job. Immutable; progress shows in `state` on refresh. |

## Data sources

| Name | Notes |
| --- | --- |
| `fireworks_deployment` | An existing deployment. |
| `fireworks_model` | A model, such as a fine-tune's output model. |

Not implemented: autoscaling policy blocks, dataset file upload,
reinforcement fine-tuning jobs.

## Development

Go as pinned in [`go.mod`](go.mod), Terraform 1.5+.

```bash
go build -o bin/terraform-provider-fireworks .
TF_ACC=1 go test ./...
```

```powershell
go build -o bin/terraform-provider-fireworks.exe .
$env:TF_ACC = "1"; go test ./...
```

The acceptance tests run against an in-process mock of the Fireworks API, so
they need no credentials and no GPU spend.

CI runs the build, `golangci-lint`, `govulncheck`, a `tfplugindocs` drift
check, and the acceptance tests across Terraform 1.5 through latest. `docs/`
is generated: regenerate it after a schema change or CI fails on the diff.

For a local build, put `dev_overrides` for `amrutp24/fireworks` in a
`dev.tfrc` pointing at `bin/` and set `TF_CLI_CONFIG_FILE` to it ([dev
overrides](https://developer.hashicorp.com/terraform/cli/config/config-file#development-overrides-for-provider-developers)).

Contributing: [CONTRIBUTING.md](CONTRIBUTING.md).
Security reports: [SECURITY.md](SECURITY.md).

## License

[MPL-2.0](LICENSE)
