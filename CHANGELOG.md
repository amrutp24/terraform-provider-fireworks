# Changelog

All notable changes to this provider. Versions follow [SemVer](https://semver.org).

## Unreleased

### Added
- `docs/` now generated from the provider schema with tfplugindocs; CI fails on drift
- CI: golangci-lint, govulncheck, `go mod tidy` check, and acceptance tests across a Terraform version matrix (1.5 → latest)
- Dependabot for Go modules and GitHub Actions; all CI actions pinned to commit SHAs
- Community health files (contributing, security policy, issue/PR templates)

## [0.2.0] - 2026-07-08

### Added
- `fireworks_supervised_fine_tuning_job` resource — asynchronous SFT jobs (immutable; state tracked on refresh)
- `fireworks_model` data source — look up a model, e.g. a fine-tune's output model

## [0.1.0] - 2026-07-07

### Added
- Initial release: `fireworks_deployment` resource (full CRUD, import, waits for `READY`), `fireworks_dataset` resource (immutable), `fireworks_deployment` data source
- Mock-API acceptance test suite; GoReleaser release pipeline with GPG-signed registry artifacts
