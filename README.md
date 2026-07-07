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
- `fireworks_deployment` (data source) — look up an existing deployment.

Roadmap: fine-tuning job resource, autoscaling policy block, dataset file
upload, model data source.

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

## Publishing to the Terraform Registry

One-time setup, following the
[official guide](https://developer.hashicorp.com/terraform/registry/providers/publishing):

1. **Create a public GitHub repo** named exactly `terraform-provider-fireworks`
   and push this code. The registry namespace equals the GitHub owner
   (`amrutp24`).
2. **Create a GPG signing key**: `gpg --full-generate-key` (RSA 4096), then
   export the ASCII-armored public key.
3. **Add repo secrets** `GPG_PRIVATE_KEY` and `PASSPHRASE` for the release
   workflow.
4. **Tag a release**: `git tag v0.1.0 && git push origin v0.1.0` — the
   `release` workflow builds, signs, and publishes artifacts via GoReleaser.
5. **Register on the registry**: sign in at
   [registry.terraform.io](https://registry.terraform.io) with GitHub, add
   your GPG public key, then "Publish → Provider" and select the repo.
   Future tags publish automatically.

## License

[MPL-2.0](LICENSE)
