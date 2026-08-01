# Changelog

All notable changes to this provider. Versions follow [SemVer](https://semver.org).

## Unreleased

### Security
- `google.golang.org/grpc` v1.79.3 → v1.82.1 (GO-2026-6061)

## [0.3.0] - 2026-07-11

### Added
- `docs/` now generated from the provider schema with tfplugindocs; CI fails on drift
- CI: golangci-lint, govulncheck, `go mod tidy` check, and acceptance tests across a Terraform version matrix (1.5 → latest)
- Dependabot for Go modules and GitHub Actions; all CI actions pinned to commit SHAs
- Community health files (contributing, security policy, issue/PR templates)

### Security
- Go toolchain 1.26.2 → 1.26.5 and `golang.org/x/net` v0.52.0 → v0.57.0, clearing all govulncheck findings (crypto/tls ECH leak, net/textproto, crypto/x509, x/net idna and HTTP/2 issues)

## [0.2.0] - 2026-07-08

### Added
- `fireworks_supervised_fine_tuning_job` resource — asynchronous SFT jobs (immutable; state tracked on refresh)
- `fireworks_model` data source — look up a model, e.g. a fine-tune's output model

## [0.1.0] - 2026-07-07

### Added
- Initial release: `fireworks_deployment` resource (full CRUD, import, waits for `READY`), `fireworks_dataset` resource (immutable), `fireworks_deployment` data source
- Mock-API acceptance test suite; GoReleaser release pipeline with GPG-signed registry artifacts
