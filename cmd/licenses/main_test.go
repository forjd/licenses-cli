package main

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	// Never depend on the machine's git config.
	gitUserName = func() string { return "" }
	os.Exit(m.Run())
}

// runArgs runs the CLI and returns stdout, stderr and the error.
func runArgs(args ...string) (string, string, error) {
	var stdout, stderr bytes.Buffer
	err := run(args, &stdout, &stderr)
	return stdout.String(), stderr.String(), err
}

// placeholderRE matches template placeholders such as [year], [yyyy] or <name of author>.
var placeholderRE = regexp.MustCompile(`\[[a-z ]+\]|<[a-z ]+>`)

// appendixPlaceholders are upstream "how to apply this license" examples that stay as text.
var appendixPlaceholders = map[string]bool{
	"Copyright [yyyy] [name of copyright owner]":        true,
	"Copyright (C) <year>  <name of author>":            true,
	"<program>  Copyright (C) <year>  <name of author>": true,
}

func TestEveryLicenseRenders(t *testing.T) {
	for _, l := range catalog {
		body, err := render(l, "Jane Doe", "2026")
		if err != nil {
			t.Fatalf("%s: %v", l.ID, err)
		}
		if needsName(l) && (!strings.Contains(body, "Jane Doe") || !strings.Contains(body, "2026")) {
			t.Errorf("%s: name or year missing from output", l.ID)
		}
		for _, line := range strings.Split(body, "\n") {
			if placeholderRE.MatchString(line) && !appendixPlaceholders[strings.TrimSpace(line)] {
				t.Errorf("%s: unreplaced placeholder: %q", l.ID, strings.TrimSpace(line))
			}
		}
	}
}

func TestCatalogMatchesEmbeddedFiles(t *testing.T) {
	files, err := fs.Glob(texts, "licenses/*.txt")
	if err != nil {
		t.Fatal(err)
	}
	inCatalog := map[string]bool{}
	for _, l := range catalog {
		inCatalog["licenses/"+l.File] = true
		if _, err := texts.ReadFile("licenses/" + l.File); err != nil {
			t.Errorf("%s: %v", l.ID, err)
		}
	}
	for _, f := range files {
		if !inCatalog[f] {
			t.Errorf("%s is embedded but not in the catalog", f)
		}
	}
}

func TestRenderWithoutName(t *testing.T) {
	mit, _ := find("mit")
	body, err := render(mit, "", "2020")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(body, "Copyright (c) 2020 [fullname]") {
		t.Errorf("unexpected output:\n%s", body)
	}
	apache, _ := find("apache-2.0")
	if !needsName(mit) || needsName(apache) {
		t.Error("needsName: want true for MIT, false for Apache-2.0")
	}
}

func TestMITStdout(t *testing.T) {
	for _, args := range [][]string{
		{"mit", "-n", "Jane Doe", "-y", "2020", "-o", "-"},
		{"-n", "Jane Doe", "-y", "2020", "-o", "-", "mit"},
		{"-n", "Jane Doe", "mit", "-y", "2020", "-o", "-"},
		{"-n", "Jane Doe", "-y", "2020", "-o", "-", "--", "mit"},
	} {
		out, _, err := runArgs(args...)
		if err != nil {
			t.Fatalf("%v: %v", args, err)
		}
		if !strings.Contains(out, "Copyright (c) 2020 Jane Doe") {
			t.Errorf("%v: unexpected output:\n%s", args, out)
		}
	}
}

func TestDefaultYear(t *testing.T) {
	out, _, err := runArgs("mit", "-n", "X", "-o", "-")
	if err != nil {
		t.Fatal(err)
	}
	if want := strconv.Itoa(time.Now().Year()); !strings.Contains(out, "Copyright (c) "+want+" X") {
		t.Errorf("want year %s in:\n%s", want, out)
	}
}

func TestYearValidation(t *testing.T) {
	for _, y := range []string{"2024", "2019-2024", "2019, 2024", "2019,2021-2024"} {
		if _, _, err := runArgs("mit", "-n", "X", "-o", "-", "-y", y); err != nil {
			t.Errorf("-y %q: %v", y, err)
		}
	}
	for _, y := range []string{"", "24", "2020\nEVIL", "next year"} {
		if _, _, err := runArgs("mit", "-n", "X", "-o", "-", "-y", y); err == nil || !strings.Contains(err.Error(), "invalid year") {
			t.Errorf("-y %q: want invalid year error, got %v", y, err)
		}
	}
}

func TestFind(t *testing.T) {
	for id, want := range map[string]string{
		"mit":               "MIT",
		"MIT":               "MIT",
		"apache-2.0":        "Apache-2.0",
		"Apache-2.0":        "Apache-2.0",
		"bSd-3-ClAuSe":      "BSD-3-Clause",
		"0bsd":              "0BSD",
		"isc":               "ISC",
		"unlicense":         "Unlicense",
		"GPL-3.0-only":      "GPL-3.0",
		"gpl-3.0-or-later":  "GPL-3.0",
		"LGPL-3.0-only":     "LGPL-3.0",
		"AGPL-3.0-or-later": "AGPL-3.0",
	} {
		l, ok := find(id)
		if !ok || l.ID != want {
			t.Errorf("find(%q) = %q, %v; want %q", id, l.ID, ok, want)
		}
	}
	for _, id := range []string{"", "apache", "gpl", "mit.txt"} {
		if _, ok := find(id); ok {
			t.Errorf("find(%q) should fail", id)
		}
	}
}

func TestErrors(t *testing.T) {
	for _, tc := range []struct {
		args []string
		want string
	}{
		{[]string{"nope"}, `unknown license "nope"`},
		{[]string{}, "missing license id"},
		{[]string{"mit", "-zzz"}, "run 'licenses -help'"},
		{[]string{"mit", "-n"}, "flag needs an argument"},
		{[]string{"mit", "-n", "Jane", "Doe"}, `unexpected argument "Doe"`},
		{[]string{"mit", "apache-2.0"}, `unexpected argument "apache-2.0"`},
		{[]string{"mit", "-n", ""}, "-n is empty"},
		{[]string{"mit", "-n", "   "}, "-n is empty"},
		{[]string{"mit", "-n", "X", "-o", ""}, "-o is empty"},
	} {
		_, _, err := runArgs(tc.args...)
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%q: want error containing %q, got %v", tc.args, tc.want, err)
		}
	}
}

func TestMissingIDPrintsUsage(t *testing.T) {
	out, errOut, _ := runArgs()
	if out != "" || !strings.Contains(errOut, "Usage:") {
		t.Errorf("want usage on stderr only; stdout=%q stderr=%q", out, errOut)
	}
}

func TestWriteAndOverwrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "LICENSE")

	out, _, err := runArgs("mit", "-n", "First", "-o", path)
	if err != nil {
		t.Fatal(err)
	}
	if want := "Wrote " + path + " (MIT)\n"; out != want {
		t.Errorf("stdout = %q, want %q", out, want)
	}

	_, _, err = runArgs("mit", "-n", "Second", "-o", path)
	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("want already exists error, got %v", err)
	}
	if b, _ := os.ReadFile(path); !strings.Contains(string(b), "First") {
		t.Error("file changed without -f")
	}

	for _, f := range []string{"-f", "--force"} {
		if _, _, err := runArgs("isc", "-n", "Third"+f, "-o", path, f); err != nil {
			t.Fatalf("%s: %v", f, err)
		}
		if b, _ := os.ReadFile(path); !strings.Contains(string(b), "Third"+f) {
			t.Errorf("%s did not overwrite", f)
		}
	}
	if entries, _ := os.ReadDir(filepath.Dir(path)); len(entries) != 1 {
		t.Errorf("temp files left behind: %v", entries)
	}
}

func TestForceKeepsFileMode(t *testing.T) {
	path := filepath.Join(t.TempDir(), "LICENSE")
	if err := os.WriteFile(path, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := runArgs("mit", "-n", "X", "-o", path, "-f"); err != nil {
		t.Fatal(err)
	}
	fi, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode().Perm() != 0o600 {
		t.Errorf("mode = %v, want 0600", fi.Mode().Perm())
	}
}

func TestDanglingSymlinkNotFollowed(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "elsewhere")
	link := filepath.Join(dir, "LICENSE")
	if err := os.Symlink(target, link); err != nil {
		t.Skip("symlinks unsupported:", err)
	}
	_, _, err := runArgs("mit", "-n", "X", "-o", link)
	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("want already exists error, got %v", err)
	}
	if _, err := os.Lstat(target); err == nil {
		t.Error("wrote through a dangling symlink")
	}

	// -f replaces the link itself, not the file it points to.
	if _, _, err := runArgs("mit", "-n", "X", "-o", link, "-f"); err != nil {
		t.Fatal(err)
	}
	if fi, _ := os.Lstat(link); fi.Mode()&os.ModeSymlink != 0 {
		t.Error("-f left the symlink in place")
	}
	if _, err := os.Lstat(target); err == nil {
		t.Error("-f wrote through the symlink")
	}
}

func TestDirectoryTarget(t *testing.T) {
	dir := t.TempDir()
	for _, args := range [][]string{{"mit", "-n", "X", "-o", dir}, {"mit", "-n", "X", "-o", dir, "-f"}} {
		_, _, err := runArgs(args...)
		if err == nil || !strings.Contains(err.Error(), "is a directory") {
			t.Errorf("%v: want is a directory error, got %v", args, err)
		}
	}
}

func TestWarnings(t *testing.T) {
	path := filepath.Join(t.TempDir(), "LICENSE")
	for _, tc := range []struct {
		args []string
		want string
	}{
		{[]string{"mit", "-o", "-"}, "no name given"},
		{[]string{"mit", "-o", path}, "no name given"},
		{[]string{"apache-2.0", "-n", "X", "-o", "-"}, "-n was ignored"},
		{[]string{"mit", "-n", "X", "-o", "-"}, ""},
		{[]string{"apache-2.0", "-o", "-"}, ""},
	} {
		_, errOut, err := runArgs(tc.args...)
		if err != nil {
			t.Fatalf("%q: %v", tc.args, err)
		}
		if (tc.want == "" && errOut != "") || !strings.Contains(errOut, tc.want) {
			t.Errorf("%q: stderr = %q, want %q", tc.args, errOut, tc.want)
		}
	}
}

func TestGitUserNameDefault(t *testing.T) {
	old := gitUserName
	gitUserName = func() string { return "Git User" }
	defer func() { gitUserName = old }()
	out, _, err := runArgs("mit", "-o", "-")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Git User") {
		t.Errorf("git user.name not used:\n%s", out)
	}
}

func TestList(t *testing.T) {
	for _, cmd := range []string{"list", "ls"} {
		out, _, err := runArgs(cmd)
		if err != nil {
			t.Fatal(err)
		}
		lines := strings.Split(strings.TrimSpace(out), "\n")
		if len(lines) != len(catalog) || !strings.HasPrefix(lines[0], "0BSD ") {
			t.Errorf("%s: unexpected output:\n%s", cmd, out)
		}
	}
}

func TestVersion(t *testing.T) {
	out, _, err := runArgs("-version")
	if err != nil || out != version+"\n" {
		t.Errorf("got %q, %v", out, err)
	}
}

func TestHelpShowsBanner(t *testing.T) {
	for _, args := range [][]string{{"-help"}, {"--help"}, {"-h"}, {"help"}} {
		out, errOut, err := runArgs(args...)
		if err != nil {
			t.Fatalf("%v: %v", args, err)
		}
		if !strings.HasPrefix(out, banner+version+"\n") || !strings.Contains(out, "Usage:") || errOut != "" {
			t.Errorf("%v: unexpected help:\n%s", args, out)
		}
	}
}
