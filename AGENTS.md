# AGENTS.md

Niwashi (`nwsctl`) is a Go CLI for infrastructure provisioning and configuration management. Module `github.com/sony/niwashi`, requires Go 1.24+.

## Build

```bash
make build        # builds ./nwsctl
```

## Code Style

Follow standard Go conventions and keep code `gofmt`-formatted. Code must pass `golangci-lint` (`.golangci.yml`: govet, errcheck, staticcheck, gosec, gocognit).

## Lint

```bash
make setup         # one-time: installs golangci-lint into ./bin (requires network access)
make lint           # golangci-lint run ./internal/... ./cmd/...
```

## Test

```bash
make unit-test      # go test ./internal/... ./cmd/... --cover
make e2e-test        # go test ./test/e2e/... -tags e2e (starts a Docker container per test; requires a running Docker daemon and network access to pull the SSH server image)
make check            # lint + unit-test + e2e-test
```

Run `make check` before submitting changes.

## Docs

```bash
git submodule update --init   # one-time: fetches the hugo-book theme
make doc                       # builds the Hugo user guide under docs/user-guide/ (requires Hugo)
```

## Changes and PRs

Keep each change reviewable in scope. If a change is likely to grow large, consider splitting it into multiple PRs before starting.

## Layout

- `cmd/nwsctl/` — CLI entrypoint
- `internal/` — implementation
- `test/e2e/` — end-to-end tests (build tag `e2e`)
- `examples/` — sample recipes and state files
- `docs/user-guide/` — Hugo user documentation (English/Japanese)
