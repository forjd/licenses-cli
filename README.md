# licenses-cli

Generate open source LICENSE files from a single, dependency-free binary.

```sh
licenses list                  # show available licenses
licenses mit -n "Jane Doe"     # write LICENSE (year defaults to now)
licenses apache-2.0 -o -       # print to stdout
licenses mit -f                # overwrite an existing LICENSE
```

If `-n` is omitted, the name comes from `git config user.name`.

Included: 0BSD, AGPL-3.0, Apache-2.0, BSD-2-Clause, BSD-3-Clause, GPL-3.0,
ISC, LGPL-3.0, MIT, MPL-2.0, Unlicense. License texts are from
[choosealicense.com](https://choosealicense.com) and embedded at build time.

## Install

```sh
go install github.com/forjd/licenses-cli@latest
```

Or build from source (Go 1.27+, or `mise install`):

```sh
make build   # bin/licenses
make dist    # cross-compiled binaries in dist/
make test
```

## License

[MIT](LICENSE)
