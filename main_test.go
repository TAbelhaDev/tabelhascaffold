package main

import (
	"bytes"
	"strings"
	"testing"
	"text/template"
)

func TestHumanizeTitle(t *testing.T) {
	cases := map[string]string{
		"tabelakanban": "Tabelakanban",
		"djobs":        "Djobs",
		"tabelatuiui":  "Tabelatuiui",
		"my-cool-tool": "MyCoolTool",
		"my_tool":      "MyTool",
	}
	for in, want := range cases {
		if got := humanizeTitle(in); got != want {
			t.Errorf("humanizeTitle(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestUpdateREADMEBadgesInsert(t *testing.T) {
	readme := "# My Tool\n\nSome description.\n"
	block := badgeBlock(project{Name: "my-tool", Org: "TabelaDev"})
	got := updateREADMEBadges(readme, block)
	if !strings.Contains(got, block) {
		t.Fatalf("badge block not inserted:\n%s", got)
	}
	if !strings.HasPrefix(got, "# My Tool\n") {
		t.Fatalf("heading moved:\n%s", got)
	}
}

func TestUpdateREADMEBadgesReplace(t *testing.T) {
	readme := "# My Tool\n\n[![Go Version](https://img.shields.io/github/go-mod/go-version/ianptkcs/old?style=flat-square)](go.mod)\n[![ko-fi](https://ko-fi.com/img/githubbutton_sm.svg)](https://ko-fi.com/old)\n\nBody.\n"
	block := badgeBlock(project{Name: "my-tool", Org: "TabelaDev"})
	got := updateREADMEBadges(readme, block)
	if strings.Contains(got, "ianptkcs/old") {
		t.Fatalf("old badge not replaced:\n%s", got)
	}
	if strings.Count(got, "[![ko-fi]") != 1 {
		t.Fatalf("ko-fi should appear exactly once, got:\n%s", got)
	}
	if !strings.Contains(got, "Body.") {
		t.Fatalf("body clobbered:\n%s", got)
	}
}

func TestUpdateREADMEBadgesIdempotent(t *testing.T) {
	block := badgeBlock(project{Name: "tabelakanban", Org: "TabelaDev"})
	once := updateREADMEBadges("# TabelaKanban\n\nBody.\n", block)
	twice := updateREADMEBadges(once, block)
	if once != twice {
		t.Fatalf("second pass changed output:\n---once---\n%s\n---twice---\n%s", once, twice)
	}
}

func TestBadgeBlockKoFiBelow(t *testing.T) {
	block := badgeBlock(project{Name: "x", Org: "TabelaDev"})
	lines := strings.Split(block, "\n")
	// The ko-fi button must be on its own line at the end, below the badges.
	lastNonEmpty := ""
	for _, l := range lines {
		if strings.TrimSpace(l) != "" {
			lastNonEmpty = l
		}
	}
	if !strings.Contains(lastNonEmpty, "ko-fi.com") {
		t.Fatalf("ko-fi not last:\n%s", block)
	}
}

func TestRenderReleaseWorkflow(t *testing.T) {
	tmpl, err := template.New("t").Parse(releaseYAML)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, project{Name: "testeapp"}); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	want := `BIN="testeapp-${{ matrix.goos }}-${{ matrix.goarch }}"`
	if !strings.Contains(out, want) {
		t.Fatalf("expected %q in:\n%s", want, out)
	}
	if strings.Contains(out, "{{.Name}}") {
		t.Fatalf("unrendered template var:\n%s", out)
	}
}
