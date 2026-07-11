# terraform-provider-fireworks

Terraform provider for [Fireworks AI](https://fireworks.ai): manage GPU
inference deployments and datasets as infrastructure-as-code.

As of July 2026 there is no other Fireworks provider on the Terraform
Registry — deployments are otherwise managed with `firectl` or the console.

```terraform
resource "fireworks_deployment" "llama" {
  base_model        = "accounts/fireworks/models/llama-v3p1-8b-instruct"
  min_replica_count = 0 # scale to zero when idle
  max_replica_count = 2
  accelerator_type  = "NVIDIA_H100_80GB"
}
```

## Resources and data sources

- `fireworks_deployment` (resource) — on-demand GPU inference deployment:
  full CRUD, import, waits for `READY` state.
- `fireworks_dataset` (resource) — fine-tuning/eval dataset (immutable,
  changes force replacement).
- `fireworks_supervised_fine_tuning_job` (resource) — asynchronous SFT job
  (immutable; track progress via `state` on refresh).
- `fireworks_deployment` (data source) — look up an existing deployment.
- `fireworks_model` (data source) — look up a model, e.g. a fine-tune's
  output model.

Roadmap: autoscaling policy block, dataset file upload, reinforcement
fine-tuning jobs.

## Authentication

| Provider attribute | Environment variable    |
|--------------------|-------------------------|
| `api_key`          | `FIREWORKS_API_KEY`     |
| `account_id`       | `FIREWORKS_ACCOUNT_ID`  |

## Development

```powershell
go build -o bin/terraform-provider-fireworks.exe .

# Acceptance tests run against an in-process mock of the Fireworks API,
# so no credentials or GPU spend are needed:
$env:TF_ACC = "1"
go test ./... -v
```

To run the example against your real account with a local build, create a
`dev.tfrc` with `dev_overrides` pointing at `bin/` (see HashiCorp's
[dev overrides docs](https://developer.hashicorp.com/terraform/cli/config/config-file#development-overrides-for-provider-developers)).

### Verification status

- `fireworks_dataset`: full lifecycle (create, refresh, destroy) verified
  against the production API.
- `fireworks_deployment`: covered by mock acceptance tests; the schema
  follows the documented API. Not yet exercised against production, since
  creating any deployment requires a payment method on the account.
- `fireworks_supervised_fine_tuning_job` and `fireworks_model`: covered by
  mock acceptance tests; running real fine-tuning jobs incurs training
  costs, so they are not exercised against production.

## License

[MPL-2.0](LICENSE)
