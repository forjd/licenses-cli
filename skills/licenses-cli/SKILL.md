---
name: licenses-cli
description: Add or replace an open source LICENSE file in a project using the `licenses` CLI. Use when the user asks to add a license, license a repo, pick between MIT/Apache/GPL/BSD/etc, change a project's license, or fix a missing or placeholder LICENSE file.
---

# licenses CLI

`licenses` writes standard license texts (from choosealicense.com) with the year and copyright holder filled in. Prefer it over typing license text by hand: hand-written license text drifts from the official wording.

## Check it is installed

```sh
licenses -version
```

If the command is missing, install it:

- macOS / Linux: `curl -fsSL https://raw.githubusercontent.com/forjd/licenses-cli/main/install.sh | sh`
- Windows (PowerShell): `irm https://raw.githubusercontent.com/forjd/licenses-cli/main/install.ps1 | iex`
- With Go: `go install github.com/forjd/licenses-cli/cmd/licenses@latest`

## Commands

```sh
licenses list                          # IDs and names of every license
licenses <id> -n "<holder>"            # write ./LICENSE
licenses <id> -n "<holder>" -y 2024    # set the year (default: current year)
licenses <id> -o LICENSE.md            # different output path
licenses <id> -o -                     # print to stdout instead of writing
licenses <id> -f                       # overwrite an existing file
```

IDs are SPDX identifiers and are not case-sensitive: `0BSD`, `AGPL-3.0`, `Apache-2.0`, `BSD-2-Clause`, `BSD-3-Clause`, `GPL-3.0`, `ISC`, `LGPL-3.0`, `MIT`, `MPL-2.0`, `Unlicense`. Run `licenses list` if unsure; the set may grow.

## Workflow

1. **Pick the license.** Use the one the user named. If they did not name one, ask; do not choose for them. If they want a suggestion: MIT for simple permissive, Apache-2.0 for permissive with a patent grant, GPL-3.0 to require derivatives stay open, MPL-2.0 for file-level copyleft.
2. **Pick the holder.** Use the name the user gave. Otherwise check, in order: an existing LICENSE's copyright line, `author` in package.json / Cargo.toml / pyproject.toml, then the git `user.name` (the CLI uses this automatically when `-n` is omitted). Confirm with the user if it is ambiguous, for example a company vs a person.
3. **Check for an existing file.** Look for `LICENSE`, `LICENSE.md`, `LICENSE.txt`, `COPYING`. The CLI refuses to overwrite without `-f`. Only pass `-f` when the user asked to replace the license, and say so in your reply.
4. **Write it** from the repo root: `licenses mit -n "Jane Doe"`.
5. **Keep metadata in sync.** Update the license field in package manifests (`"license": "MIT"` in package.json, `license = "MIT"` in Cargo.toml / pyproject.toml) and any license badge or section in the README, using the same SPDX ID.
6. **Verify:** `head -5 LICENSE` shows the right name, year and holder, with no `[year]` or `[fullname]` left.

## Notes

- Changing an existing project's license can have legal effects for past contributions. If the project has outside contributors, point that out before replacing the license.
- Apache-2.0, GPL, LGPL, AGPL, MPL and the Unlicense have no holder placeholder in the text itself, so `-n` has no effect on them. That is expected.
- Dual licensing (e.g. MIT OR Apache-2.0): write each to its own file, `licenses mit -o LICENSE-MIT` and `licenses apache-2.0 -o LICENSE-APACHE`.
