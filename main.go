// Command licenses writes open source license files.
package main

import (
	"embed"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
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
	ID    string // SPDX identifier
	Title string
	File  string
}

var catalog = []license{
	{"0BSD", "BSD Zero Clause License", "0bsd.txt"},
	{"AGPL-3.0", "GNU Affero General Public License v3.0", "agpl-3.0.txt"},
	{"Apache-2.0", "Apache License 2.0", "apache-2.0.txt"},
	{"BSD-2-Clause", `BSD 2-Clause "Simplified" License`, "bsd-2-clause.txt"},
	{"BSD-3-Clause", `BSD 3-Clause "New" or "Revised" License`, "bsd-3-clause.txt"},
	{"GPL-3.0", "GNU General Public License v3.0", "gpl-3.0.txt"},
	{"ISC", "ISC License", "isc.txt"},
	{"LGPL-3.0", "GNU Lesser General Public License v3.0", "lgpl-3.0.txt"},
	{"MIT", "MIT License", "mit.txt"},
	{"MPL-2.0", "Mozilla Public License 2.0", "mpl-2.0.txt"},
	{"Unlicense", "The Unlicense", "unlicense.txt"},
}

const usage = `licenses - generate a LICENSE file

Usage:
  licenses <id> [flags]    write a license (e.g. licenses mit -n "Jane Doe")
  licenses list            show available licenses

Flags:
  -n, -name string   copyright holder (default: git config user.name)
  -y, -year string   copyright year (default: current year)
  -o, -out string    output file, "-" for stdout (default "LICENSE")
  -f, -force         overwrite an existing file
  -version           print version
`

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "licenses:", err)
		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer) error {
	// Allow "licenses mit -n Dan": take the positional id first, then parse flags.
	var id string
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		id, args = args[0], args[1:]
	}

	fs := flag.NewFlagSet("licenses", flag.ContinueOnError)
	fs.Usage = func() { fmt.Fprint(os.Stderr, usage) }
	var name, year, out string
	var force, showVersion bool
	for _, n := range []string{"n", "name"} {
		fs.StringVar(&name, n, "", "")
	}
	for _, n := range []string{"y", "year"} {
		fs.StringVar(&year, n, strconv.Itoa(time.Now().Year()), "")
	}
	for _, n := range []string{"o", "out"} {
		fs.StringVar(&out, n, "LICENSE", "")
	}
	for _, n := range []string{"f", "force"} {
		fs.BoolVar(&force, n, false, "")
	}
	fs.BoolVar(&showVersion, "version", false, "")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if id == "" && fs.NArg() > 0 {
		id = fs.Arg(0)
	}

	switch {
	case showVersion:
		fmt.Fprintln(stdout, version)
		return nil
	case id == "":
		fmt.Fprint(os.Stderr, usage)
		return errors.New("missing license id")
	case id == "list" || id == "ls":
		return list(stdout)
	case id == "help":
		fmt.Fprint(stdout, usage)
		return nil
	}

	lic, ok := find(id)
	if !ok {
		return fmt.Errorf("unknown license %q (run 'licenses list')", id)
	}
	if name == "" {
		name = gitUserName()
	}
	body, err := render(lic, name, year)
	if err != nil {
		return err
	}

	if out == "-" {
		_, err = io.WriteString(stdout, body)
		return err
	}
	if !force {
		if _, err := os.Stat(out); err == nil {
			return fmt.Errorf("%s already exists (use -f to overwrite)", out)
		}
	}
	if err := os.WriteFile(out, []byte(body), 0o644); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "Wrote %s (%s)\n", out, lic.ID)
	if name == "" && needsName(lic) {
		fmt.Fprintln(os.Stderr, "warning: no name given; use -n to set the copyright holder")
	}
	return nil
}

func find(id string) (license, bool) {
	for _, l := range catalog {
		if strings.EqualFold(l.ID, id) || strings.EqualFold(strings.TrimSuffix(l.File, ".txt"), id) {
			return l, true
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
func gitUserName() string {
	out, err := exec.Command("git", "config", "user.name").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
