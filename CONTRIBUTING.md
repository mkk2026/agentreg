# Contributing to agentreg

Thanks for taking a look. agentreg is early (v0.x) and feedback, bug reports, and
PRs are all welcome.

## Ways to help

- **Try it and tell us what broke.** The most useful contribution right now is
  honest feedback from running it against real agents. Open an issue.
- **Report a bug** using the bug template.
- **Propose a feature** using the feature template — but check the
  [Roadmap](README.md#roadmap) first; some things are deliberately deferred.
- **Send a PR** for a bug fix or a small, focused improvement.

## Development

Requires Go 1.26+.

```bash
git clone https://github.com/mkk2026/agentreg
cd agentreg
go test ./...        # run the tests
go vet ./...         # static checks
go build -o agentctl # build the binary
```

## Pull requests

- Keep PRs focused — one logical change per PR.
- Add or update tests for behavior changes. CI runs `go vet` and `go test -race`
  on every PR and must pass.
- Run `gofmt -w .` before committing.
- For anything larger than a bug fix, open an issue first so we can agree on the
  approach before you invest the time.

## Design principle

agentreg wins on being small and delightful, not on having every feature. New
functionality has to earn its place. When in doubt, keep the core lean and put
richer behavior behind the existing seams (the `Verifier` interface, the
`Source`/`labels` fields).

## Questions

Open a [Discussion](https://github.com/mkk2026/agentreg/discussions) or an issue.
