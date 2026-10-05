// Command licenses writes open source license files.
package main

import (
	"embed"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

//go:embed licenses/*.txt
var texts embed.FS

// version is set at build time with -ldflags "-X main.version=...".
var version = "dev"

type license struct {
	ID      string // SPDX identifier
	Title   string
	File    string
	Aliases []string // other SPDX IDs with the same text
}

var catalog = []license{
	{"0BSD", "BSD Zero Clause License", "0bsd.txt", nil},
	{"AGPL-3.0", "GNU Affero General Public License v3.0", "agpl-3.0.txt", []string{"AGPL-3.0-only", "AGPL-3.0-or-later"}},
	{"Apache-2.0", "Apache License 2.0", "apache-2.0.txt", nil},
	{"BSD-2-Clause", `BSD 2-Clause "Simplified" License`, "bsd-2-clause.txt", nil},
	{"BSD-3-Clause", `BSD 3-Clause "New" or "Revised" License`, "bsd-3-clause.txt", nil},
	{"GPL-3.0", "GNU General Public License v3.0", "gpl-3.0.txt", []string{"GPL-3.0-only", "GPL-3.0-or-later"}},
	{"ISC", "ISC License", "isc.txt", nil},
	{"LGPL-3.0", "GNU Lesser General Public License v3.0", "lgpl-3.0.txt", []string{"LGPL-3.0-only", "LGPL-3.0-or-later"}},
	{"MIT", "MIT License", "mit.txt", nil},
	{"MPL-2.0", "Mozilla Public License 2.0", "mpl-2.0.txt", nil},
	{"Unlicense", "The Unlicense", "unlicense.txt", nil},
}

// banner is figlet -f small "licenses".
const banner = ` _ _
| (_)__ ___ _ _  ___ ___ ___
| | / _/ -_) ' \(_-</ -_|_-<
|_|_\__\___|_||_/__/\___/__/
`

const usage = `licenses - generate a LICENSE file

Usage:
  licenses <id> [flags]    write a license (e.g. licenses mit -n "Jane Doe")
  licenses list            show available licenses (alias: ls)
  licenses help            show this help

Flags:
  -n, -name string   copyright holder (default: git config user.name)
  -y, -year string   copyright year or range, e.g. 2019-2024 (default: current year)
  -o, -out string    output file, "-" for stdout (default "LICENSE")
  -f, -force         overwrite an existing file
  -version           print version
  -h, -help          show this help
`

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "licenses:", err)
		os.Exit(1)
	}
}

// yearRE accepts a year, a range, or a comma-separated list of either: 2024, 2019-2024, 2019, 2024.
var yearRE = regexp.MustCompile(`^[0-9]{4}(-[0-9]{4})?(, ?[0-9]{4}(-[0-9]{4})?)*$`)

func run(args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("licenses", flag.ContinueOnError)
	// Help and errors are printed by run, not by the flag package.
	flags.SetOutput(io.Discard)
	flags.Usage = func() {}
	var name, year, out string
	var force, showVersion bool
	for _, n := range []string{"n", "name"} {
		flags.StringVar(&name, n, "", "")
	}
	for _, n := range []string{"y", "year"} {
		flags.StringVar(&year, n, strconv.Itoa(time.Now().Year()), "")
	}
	for _, n := range []string{"o", "out"} {
		flags.StringVar(&out, n, "LICENSE", "")
	}
	for _, n := range []string{"f", "force"} {
		flags.BoolVar(&force, n, false, "")
	}
	flags.BoolVar(&showVersion, "version", false, "")

	// The flag package stops at the first positional argument, so parse in a
	// loop to allow flags on both sides of the id: "licenses -n Dan mit -f".
	var positional []string
	for {
		if err := flags.Parse(args); err != nil {
			if errors.Is(err, flag.ErrHelp) {
				printHelp(stdout)
				return nil
			}
			return fmt.Errorf("%w (run 'licenses -help')", err)
		}
		if flags.NArg() == 0 {
			break
		}
		positional = append(positional, flags.Arg(0))
		args = flags.Args()[1:]
	}
	if len(positional) > 1 {
		return fmt.Errorf("unexpected argument %q (quote names with spaces: -n \"Jane Doe\")", positional[1])
	}
	var id string
	if len(positional) == 1 {
		id = positional[0]
	}

	switch {
	case showVersion:
		fmt.Fprintln(stdout, version)
		return nil
	case id == "":
		fmt.Fprint(stderr, usage)
		return errors.New("missing license id")
	case id == "list" || id == "ls":
		return list(stdout)
	case id == "help":
		printHelp(stdout)
		return nil
	}

	lic, ok := find(id)
	if !ok {
		return fmt.Errorf("unknown license %q (run 'licenses list')", id)
	}
	nameSet := false
	flags.Visit(func(f *flag.Flag) { nameSet = nameSet || f.Name == "n" || f.Name == "name" })
	name = strings.TrimSpace(name)
	switch {
	case nameSet && name == "":
		return errors.New("-n is empty")
	case !nameSet:
		name = gitUserName()
	}
	if !yearRE.MatchString(year) {
		return fmt.Errorf("invalid year %q (e.g. 2024 or 2019-2024)", year)
	}
	if out == "" {
		return errors.New("-o is empty (use - for stdout)")
	}
	body, err := render(lic, name, year)
	if err != nil {
		return err
	}

	if out == "-" {
		_, err = io.WriteString(stdout, body)
	} else if err = writeFile(out, body, force); err == nil {
		fmt.Fprintf(stdout, "Wrote %s (%s)\n", out, lic.ID)
	}
	if err != nil {
		return err
	}
	switch {
	case needsName(lic) && name == "":
		fmt.Fprintln(stderr, "warning: no name given; use -n to set the copyright holder")
	case !needsName(lic) && nameSet:
		fmt.Fprintf(stderr, "warning: %s has no copyright holder line; -n was ignored\n", lic.ID)
	}
	return nil
}

// writeFile writes body to path. Without force it never replaces anything,
// including a dangling symlink. With force it writes a temp file and renames
// it into place, so a failed write leaves the old file intact.
func writeFile(path, body string, force bool) error {
	if fi, err := os.Stat(path); err == nil && fi.IsDir() {
		return fmt.Errorf("%s is a directory", path)
	}
	if !force {
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if errors.Is(err, fs.ErrExist) {
			return fmt.Errorf("%s already exists (use -f to overwrite)", path)
		}
		if err != nil {
			return err
		}
		_, err = f.WriteString(body)
		if cerr := f.Close(); err == nil {
			err = cerr
		}
		if err != nil {
			os.Remove(path)
		}
		return err
	}

	mode := fs.FileMode(0o644)
	if fi, err := os.Lstat(path); err == nil && fi.Mode().IsRegular() {
		mode = fi.Mode().Perm()
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".licenses-*")
	if err != nil {
		return err
	}
	tmp := f.Name()
	_, err = f.WriteString(body)
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err == nil {
		err = os.Chmod(tmp, mode)
	}
	if err == nil {
		err = os.Rename(tmp, path)
	}
	if err != nil {
		os.Remove(tmp)
	}
	return err
}

func printHelp(w io.Writer) {
	fmt.Fprintf(w, "%s%s\n\n%s", banner, version, usage)
}

func find(id string) (license, bool) {
	for _, l := range catalog {
		if strings.EqualFold(l.ID, id) {
			return l, true
		}
		for _, a := range l.Aliases {
			if strings.EqualFold(a, id) {
				return l, true
			}
		}
	}
	return license{}, false
}

func render(l license, name, year string) (string, error) {
	b, err := texts.ReadFile("licenses/" + l.File)
	if err != nil {
		return "", err
	}
	if name == "" {
		name = "[fullname]"
	}
	return strings.NewReplacer("[year]", year, "[fullname]", name).Replace(string(b)), nil
}

func needsName(l license) bool {
	b, _ := texts.ReadFile("licenses/" + l.File)
	return strings.Contains(string(b), "[fullname]")
}

func list(w io.Writer) error {
	ls := append([]license(nil), catalog...)
	sort.Slice(ls, func(i, j int) bool { return strings.ToLower(ls[i].ID) < strings.ToLower(ls[j].ID) })
	for _, l := range ls {
		fmt.Fprintf(w, "%-14s %s\n", l.ID, l.Title)
	}
	return nil
}

// gitUserName returns git's user.name, or "" if git is missing or unset.
// It is a variable so tests can stub it.
var gitUserName = func() string {
	out, err := exec.Command("git", "config", "user.name").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
