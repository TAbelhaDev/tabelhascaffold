package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/template"
)

func TestSplitArgs(t *testing.T) {
	// release style: dir first, then --version — the case that used to break.
	dir, rest := splitArgs([]string{".", "--version", "v0.2.0"}, map[string]bool{"--version": true})
	if dir != "." || len(rest) != 2 || rest[0] != "--version" || rest[1] != "v0.2.0" {
		t.Fatalf("splitArgs(dir-first) = dir=%q rest=%v", dir, rest)
	}

	// flag first, then dir
	dir, rest = splitArgs([]string{"--version", "v0.2.0", "."}, map[string]bool{"--version": true})
	if dir != "." || len(rest) != 2 {
		t.Fatalf("splitArgs(flag-first) = dir=%q rest=%v", dir, rest)
	}

	// no dir at all → defaults to "."
	dir, rest = splitArgs([]string{"--version", "v0.1.0"}, map[string]bool{"--version": true})
	if dir != "." {
		t.Fatalf("splitArgs(no dir) dir=%q, want .", dir)
	}
}

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

func TestUpdateHeaderInsert(t *testing.T) {
	readme := "# My Tool\n\nSome description.\n"
	got := updateHeader(readme, project{Name: "my-tool", Org: "TabelaDev"})
	if !strings.HasPrefix(got, "# My Tool\n") {
		t.Fatalf("heading moved:\n%s", got)
	}
	if !strings.Contains(got, "<div align=\"center\">") {
		t.Fatalf("missing centered div:\n%s", got)
	}
	if !strings.Contains(got, "[![ko-fi]") {
		t.Fatalf("missing ko-fi button:\n%s", got)
	}
	if !strings.Contains(got, "Some description.") {
		t.Fatalf("body clobbered:\n%s", got)
	}
}

func TestUpdateHeaderReplace(t *testing.T) {
	readme := "# My Tool\n\n[![Go Version](https://img.shields.io/github/go-mod/go-version/ianptkcs/old?style=flat-square)](go.mod)\n[![ko-fi](https://ko-fi.com/img/githubbutton_sm.svg)](https://ko-fi.com/old)\n\nBody.\n"
	got := updateHeader(readme, project{Name: "my-tool", Org: "TabelaDev"})
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

func TestUpdateHeaderPreservesTagline(t *testing.T) {
	readme := "<div align=\"center\">\n\n# TabelaFin\n\n**Finanças pessoais — BYOK, sem assinatura.**\n\n[![SvelteKit](https://img.shields.io/badge/SvelteKit-Svelte-ff3e00?style=flat-square&logo=svelte&logoColor=white)](https://kit.svelte.dev)\n[![License: AGPL-3.0](https://img.shields.io/badge/license-AGPL--3.0-blue?style=flat-square)](LICENSE)\n\n[![ko-fi](https://ko-fi.com/img/githubbutton_sm.svg)](https://ko-fi.com/ianptkcs)\n\n</div>\n\n---\n\n# Body Section\n"
	got := updateHeader(readme, project{Name: "tabelafin", Org: "TabelaDev", Stack: "web"})
	if !strings.Contains(got, "**Finanças pessoais — BYOK, sem assinatura.**") {
		t.Fatalf("tagline not preserved:\n%s", got)
	}
	if !strings.Contains(got, "# Body Section") {
		t.Fatalf("body section lost:\n%s", got)
	}
}

func TestUpdateHeaderIdempotent(t *testing.T) {
	src := "<div align=\"center\">\n\n# TabelaKanban\n\n**Kanban TUI sobre markdown.**\n\n[![Go Version](https://img.shields.io/github/go-mod/go-version/TabelaDev/tabelakanban?style=flat-square&logo=go&logoColor=white&color=00ADD8)](go.mod)\n[![License: AGPL-3.0](https://img.shields.io/badge/license-AGPL--3.0-blue?style=flat-square)](LICENSE)\n[![Built with Bubble Tea](https://img.shields.io/badge/built%20with-Bubble%20Tea-ff69b4?style=flat-square)](https://github.com/charmbracelet/bubbletea)\n[![Powered by tabelatuiui](https://img.shields.io/badge/theme-tabelatuiui-d6b4f7?style=flat-square)](https://github.com/TabelaDev/tabelatuiui)\n\n[![ko-fi](https://ko-fi.com/img/githubbutton_sm.svg)](https://ko-fi.com/ianptkcs)\n\n</div>\n\n---\n\nBODY\n"
	p := project{Name: "tabelakanban", Org: "TabelaDev"}
	once := updateHeader(src, p)
	if once != src {
		t.Fatalf("canonical input changed:\n---want---\n%s\n---got---\n%s", src, once)
	}
	twice := updateHeader(once, p)
	if once != twice {
		t.Fatalf("second pass changed output")
	}
}

func TestHeaderBlockKoFiBelow(t *testing.T) {
	block := headerBlock("My Tool", "", project{Name: "x", Org: "TabelaDev"})
	lines := strings.Split(block, "\n")
	lastNonEmpty := ""
	for _, l := range lines {
		if strings.TrimSpace(l) != "" {
			lastNonEmpty = l
		}
	}
	// Only the ko-fi line and the </div> close follow the badges; the button
	// must sit on its own line above the closing div.
	if !strings.Contains(lastNonEmpty, "</div>") {
		t.Fatalf("last line should be </div>:\n%s", block)
	}
	if !strings.Contains(block, "\n\n"+kofi+"\n\n</div>") {
		t.Fatalf("ko-fi not on its own line before </div>:\n%s", block)
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

func TestSetupWebStack(t *testing.T) {
	dir := t.TempDir()
	if err := setup(dir, project{Name: "tabelafin", Org: "TabelaDev", Stack: "web"}); err != nil {
		t.Fatal(err)
	}

	ci, err := os.ReadFile(filepath.Join(dir, ".github", "workflows", "ci.yml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(ci), "oven-sh/setup-bun@v2") {
		t.Fatalf("web CI should use Bun:\n%s", ci)
	}
	if strings.Contains(string(ci), "go vet") {
		t.Fatalf("web CI must not contain Go steps:\n%s", ci)
	}

	// Web release: notes-only GitHub release, no binary matrix.
	rel, err := os.ReadFile(filepath.Join(dir, ".github", "workflows", "release.yml"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(rel), "matrix.goos") {
		t.Fatalf("web release must not build binaries:\n%s", rel)
	}
	if !strings.Contains(string(rel), "generate_release_notes: true") {
		t.Fatalf("web release should generate release notes:\n%s", rel)
	}
}

func TestWebBadgeBlock(t *testing.T) {
	block := headerBlock("TabelaFin", "", project{Name: "tabelafin", Org: "TabelaDev", Stack: "web"})
	if !strings.Contains(block, "SvelteKit") {
		t.Fatalf("missing SvelteKit badge:\n%s", block)
	}
	if !strings.Contains(block, "Cloudflare Workers") {
		t.Fatalf("missing Cloudflare badge:\n%s", block)
	}
	if !strings.Contains(block, "tabelawebui") {
		t.Fatalf("missing tabelawebui badge:\n%s", block)
	}
	if strings.Contains(block, "Bubble Tea") {
		t.Fatalf("web badges must not contain Bubble Tea:\n%s", block)
	}
	lines := strings.Split(block, "\n")
	lastNonEmpty := ""
	for _, l := range lines {
		if strings.TrimSpace(l) != "" {
			lastNonEmpty = l
		}
	}
	if !strings.Contains(lastNonEmpty, "</div>") {
		t.Fatalf("header should close with </div>:\n%s", block)
	}
	if !strings.Contains(block, "\n\n"+kofi+"\n\n</div>") {
		t.Fatalf("ko-fi not on its own line before </div>:\n%s", block)
	}
}

func TestSetupPreservesExistingFiles(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".github", "workflows"), 0o755); err != nil {
		t.Fatal(err)
	}
	custom := "# Custom CONTRIBUTING in English\n"
	customTpl := "name: Bug report\nEnglish template\n"
	englishChangelog := "# Changelog\ncustom\n"
	if err := os.WriteFile(filepath.Join(dir, "CONTRIBUTING.md"), []byte(custom), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".github", "ISSUE_TEMPLATE"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".github", "ISSUE_TEMPLATE", "bug_report.yml"), []byte(customTpl), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "CHANGELOG.md"), []byte(englishChangelog), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := setup(dir, project{Name: "myapp", Org: "TabelaDev"}); err != nil {
		t.Fatal(err)
	}

	// Custom prose must survive a re-run.
	if got, _ := os.ReadFile(filepath.Join(dir, "CONTRIBUTING.md")); string(got) != custom {
		t.Fatalf("CONTRIBUTING clobbered:\n%s", got)
	}
	if got, _ := os.ReadFile(filepath.Join(dir, ".github", "ISSUE_TEMPLATE", "bug_report.yml")); string(got) != customTpl {
		t.Fatalf("bug_report template clobbered:\n%s", got)
	}
	if got, _ := os.ReadFile(filepath.Join(dir, "CHANGELOG.md")); string(got) != englishChangelog {
		t.Fatalf("CHANGELOG clobbered:\n%s", got)
	}
	// But CI is always canonical.
	ci, err := os.ReadFile(filepath.Join(dir, ".github", "workflows", "ci.yml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(ci), "actions/checkout@v7") {
		t.Fatalf("CI not canonical:\n%s", ci)
	}
}
