# Contributing to Niwashi

Thank you for your interest in contributing to Niwashi.

## Accepting Contributions

At this time, Niwashi does not accept external contributions (pull requests) from outside Sony. We plan to open the project to external contributions in the first half of 2027.

Bug reports, questions, and feature requests via [Issues](https://github.com/sony/niwashi/issues) are welcome from anyone in the meantime.

## Reporting Issues

Before opening a new issue, please search existing issues to avoid duplicates.

When sharing logs or code snippets in an issue, please double-check that no secrets were accidentally included.

## Development

Niwashi requires Go 1.24 or later.

```bash
git clone https://github.com/sony/niwashi.git
cd niwashi
make build           # builds ./nwsctl
```

Before submitting a change, run:

```bash
make setup           # one-time: installs golangci-lint (requires network access)
make check            # lint + unit tests + e2e tests (e2e requires a running Docker daemon)
```

## Making Changes

1. Fork the repository and create a branch from `main`.
2. Make your changes, with tests where applicable.
3. Run `make check` and ensure it passes.
4. Open a pull request against `main`.
5. Address review feedback; a maintainer will merge once approved.

Keep each pull request focused and reviewable in scope. If a change is likely to grow large, split it into multiple, smaller pull requests.

## Code of Conduct

This project follows the [Code of Conduct](CODE_OF_CONDUCT.md).

## License

By contributing, you agree that your contributions will be licensed under the [Apache License 2.0](LICENSE).
