<div align="center">

# 📜 licenses

**Drop a proper LICENSE file into any project with a single command.**

One small binary with no dependencies. Works on macOS, Linux and Windows.

[![Release](https://img.shields.io/github/v/release/forjd/licenses-cli?sort=semver)](https://github.com/forjd/licenses-cli/releases/latest)
[![CI](https://github.com/forjd/licenses-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/forjd/licenses-cli/actions/workflows/ci.yml)
[![Go](https://img.shields.io/github/go-mod/go-version/forjd/licenses-cli)](go.mod)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue)](LICENSE)

</div>

```console
$ licenses mit -n "Jane Doe"
Wrote LICENSE (MIT)

$ head -3 LICENSE
MIT License

Copyright (c) 2026 Jane Doe
```

## Install

**macOS / Linux**

```sh
curl -fsSL https://raw.githubusercontent.com/forjd/licenses-cli/main/install.sh | sh
```

**Windows (PowerShell)**

```powershell
irm https://raw.githubusercontent.com/forjd/licenses-cli/main/install.ps1 | iex
```

<details>
<summary>Other ways to install</summary>

**Go**

```sh
go install github.com/forjd/licenses-cli/cmd/licenses@latest
```

**Manual:** download an archive for your platform from the [latest release](https://github.com/forjd/licenses-cli/releases/latest), extract it, and put `licenses` on your `PATH`.

**Install script options**

| Variable               | Default                                         |
| ---------------------- | ----------------------------------------------- |
| `LICENSES_VERSION`     | `latest` (or a tag such as `v0.1.0`)            |
| `LICENSES_INSTALL_DIR` | `/usr/local/bin` if writable, else `~/.local/bin`. On Windows, `%LOCALAPPDATA%\Programs\licenses` |

```sh
curl -fsSL https://raw.githubusercontent.com/forjd/licenses-cli/main/install.sh | LICENSES_VERSION=v0.1.0 sh
```

The scripts check each download against the release's `checksums.txt` before installing.

</details>

## Usage

```sh
licenses list                     # show available licenses
licenses mit -n "Jane Doe"        # write ./LICENSE
licenses apache-2.0               # holder defaults to `git config user.name`
licenses bsd-3-clause -y 2019     # set the year (default: this year)
licenses mpl-2.0 -o LICENSE.txt   # choose the output file
licenses isc -o -                 # print to stdout
licenses mit -f                   # overwrite an existing LICENSE
```

| Flag             | Description                                   |
| ---------------- | --------------------------------------------- |
| `-n`, `-name`    | Copyright holder (default: git `user.name`)   |
| `-y`, `-year`    | Copyright year (default: current year)        |
| `-o`, `-out`     | Output file, or `-` for stdout (default `LICENSE`) |
| `-f`, `-force`   | Overwrite an existing file                    |
| `-version`       | Print the version                             |

## Licenses

License IDs follow [SPDX](https://spdx.org/licenses/) and are not case-sensitive.

| ID             | License                                    |
| -------------- | ------------------------------------------ |
| `MIT`          | MIT License                                |
| `Apache-2.0`   | Apache License 2.0                         |
| `GPL-3.0`      | GNU General Public License v3.0            |
| `LGPL-3.0`     | GNU Lesser General Public License v3.0     |
| `AGPL-3.0`     | GNU Affero General Public License v3.0     |
| `MPL-2.0`      | Mozilla Public License 2.0                 |
| `BSD-2-Clause` | BSD 2-Clause "Simplified" License          |
| `BSD-3-Clause` | BSD 3-Clause "New" or "Revised" License    |
| `ISC`          | ISC License                                |
| `0BSD`         | BSD Zero Clause License                    |
| `Unlicense`    | The Unlicense                              |

The texts come from [choosealicense.com](https://choosealicense.com) and are built into the binary, so the CLI works offline.

## 🤖 Agent skill

[`skills/licenses-cli`](skills/licenses-cli/SKILL.md) teaches AI coding agents such as Claude Code to use the CLI. It covers choosing the license and holder, avoiding accidental overwrites, and updating `package.json` and other manifests to match.

```sh
# Claude Code (personal skills)
mkdir -p ~/.claude/skills/licenses-cli && curl -fsSL \
  https://raw.githubusercontent.com/forjd/licenses-cli/main/skills/licenses-cli/SKILL.md \
  -o ~/.claude/skills/licenses-cli/SKILL.md
```

You can also install it with the [`skills`](https://github.com/vercel-labs/skills) CLI: `npx skills add forjd/licenses-cli`.

## Development

Needs Go 1.27+. Run `mise install` to get the pinned toolchain.

```sh
make test    # run tests
make build   # build bin/licenses
make dist    # snapshot release for every platform (needs goreleaser)
```

To release, push a tag. GitHub Actions then runs [GoReleaser](https://goreleaser.com) and publishes the archives:

```sh
git tag v0.1.0 && git push origin v0.1.0
```

## License

[MIT](LICENSE) © Forjd
