package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestEveryLicenseRenders(t *testing.T) {
	for _, l := range catalog {
		body, err := render(l, "Jane Doe", "2026")
		if err != nil {
			t.Fatalf("%s: %v", l.ID, err)
		}
		if strings.Contains(body, "[year]") || strings.Contains(body, "[fullname]") {
			t.Errorf("%s: unreplaced placeholder", l.ID)
		}
	}
}

func TestMITStdout(t *testing.T) {
	var buf bytes.Buffer
	if err := run([]string{"mit", "-n", "Jane Doe", "-y", "2020", "-o", "-"}, &buf); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "Copyright (c) 2020 Jane Doe") {
		t.Errorf("unexpected output:\n%s", buf.String())
	}
}

func TestFindCaseInsensitive(t *testing.T) {
	for _, id := range []string{"mit", "MIT", "apache-2.0", "bsd-3-clause", "unlicense"} {
		if _, ok := find(id); !ok {
			t.Errorf("find(%q) failed", id)
		}
	}
}

func TestUnknownLicense(t *testing.T) {
	if err := run([]string{"nope"}, &bytes.Buffer{}); err == nil {
		t.Error("expected error")
	}
}
