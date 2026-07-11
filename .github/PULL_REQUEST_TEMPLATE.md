## What

<!-- One or two sentences: what changes and why. -->

## Checklist

- [ ] `go build ./...`, `go vet ./...`, and `go mod tidy -diff` pass
- [ ] `golangci-lint run` passes
- [ ] `go generate ./...` run — `docs/` regenerated, not hand-edited
- [ ] `TF_ACC=1 go test ./...` passes; new resources/data sources have mock acceptance tests
- [ ] CHANGELOG.md updated for user-visible changes
- [ ] Real-API verification status stated (verified against production / mock-only and why)

## Breaking change?

<!-- Does this change attribute names, defaults, or force replacement of
     existing resources? Say so explicitly. -->
