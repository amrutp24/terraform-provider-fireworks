# Contributing

Thanks for taking the time to contribute!

## Development setup

```bash
git clone https://github.com/amrutp24/terraform-provider-fireworks
cd terraform-provider-fireworks
go build ./...
```

## Before opening a PR

1. **Build + vet**: `go build ./... && go vet ./...`; keep `go.mod` tidy
   (`go mod tidy -diff` must be clean).
2. **Lint**: `golangci-lint run` (config in `.golangci.yml`).
3. **Docs**: `docs/` is generated — never edit it by hand. Change the schema
   `Description`s in `internal/provider/` or the files under `examples/`,
   then run `go generate ./...`. CI fails on drift.
4. **Tests**: `TF_ACC=1 go test ./... -v`. Acceptance tests run against an
   in-process mock of the Fireworks API — no credentials or GPU spend needed.
   New resources/data sources must come with mock acceptance tests.
5. **Vulnerabilities**: `go run golang.org/x/vuln/cmd/govulncheck@latest ./...`.

## Real-API changes

If your change touches request/response handling, note in the PR whether you
verified it against the production API and with which resource. Creating
deployments or fine-tuning jobs costs money — mock coverage is acceptable for
those; say so explicitly.

## Releases

Maintainer-driven: pushing a `v*` tag builds, signs, and publishes via
GoReleaser; the Terraform Registry ingests it automatically. Add a
[CHANGELOG.md](CHANGELOG.md) entry for user-visible changes.

## Reporting bugs / requesting features

Use the issue templates. For security vulnerabilities, **do not open a
public issue** — see [SECURITY.md](SECURITY.md).
