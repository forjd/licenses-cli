# AGENTS.md

`licenses` is a Go CLI that writes LICENSE files. It ships as one static binary with no third-party Go dependencies. Keep it that way: use only the standard library.

## Layout

| Path | What it is |
| --- | --- |
| `cmd/licenses/main.go` | The whole CLI: catalog, flag parsing, rendering, help banner |
| `cmd/licenses/main_test.go` | Tests |
| `cmd/licenses/licenses/*.txt` | License texts, embedded with `go:embed` |
| `skills/licenses-cli/SKILL.md` | Agent skill for using the CLI, installed with `npx skills add forjd/licenses-cli` |
| `install.sh`, `install.ps1` | One-line installers. They download from the latest GitHub release and verify `checksums.txt` |
| `.goreleaser.yaml` | Release builds and archives |
| `.github/workflows/` | `ci.yml` (push to main, PRs), `release.yml` (`v*` tags) |
| `.github/logo-*.svg`, `.github/demo.*` | README logo and demo GIF |

## Commands

The Go and GoReleaser versions are pinned in `mise.toml`. Prefix commands with `mise exec --` if they are not on your PATH.

```sh
make test    # go test ./...
make build   # bin/licenses
make dist    # goreleaser snapshot into dist/
make demo    # re-record .github/demo.gif with VHS in Docker
```

Before committing, run `gofmt -l .`, `go vet ./...` and `make test`.

## Adding a license

1. Download the text from `https://raw.githubusercontent.com/github/choosealicense.com/gh-pages/_licenses/<id>.txt`.
2. Strip the YAML front matter and save it as `cmd/licenses/licenses/<id>.txt`. Keep the `[year]` and `[fullname]` placeholders exactly as they are.
3. Add an entry to `catalog` in `main.go` with the SPDX ID and the `title` from the front matter.
4. Add it to the license table in `README.md` and the ID list in `skills/licenses-cli/SKILL.md`.
5. Run `make test`. `TestEveryLicenseRenders` fails if a placeholder is left unreplaced.

## Conventions

- Use SPDX IDs everywhere. Lookups are case-insensitive.
- Never overwrite an existing file unless `-f` is passed.
- Print errors through `run`'s returned error. `main` prefixes them with `licenses:` and exits 1. Help goes to stdout.
- If you change flags or output, update the usage text in `main.go`, the README, `SKILL.md`, and re-record the demo with `make demo`.
- Keep `install.sh` POSIX `sh` and clean under `shellcheck`.
- Archive names have no version in them (`licenses_<os>_<arch>`) so the installers can use `/releases/latest/download/`. Don't change `name_template` without updating both installers.

## CI and releases

- Workflows run on the self-hosted `forjd` runner (`runs-on: [self-hosted, forjd]`), not GitHub-hosted runners. It is Linux x64 only, so CI cross-compiles every release target instead of using an OS matrix.
- The repo is public, so CI must never run code from fork PRs on that runner. Keep the `if:` guard in `ci.yml`.
- To release, tag `main` with `vX.Y.Z` and push the tag. GoReleaser builds linux, darwin and windows for amd64 and arm64, then publishes the archives and `checksums.txt`. Only tag when asked.
