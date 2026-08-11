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
	if !strings.HasPrefix(got, "<div align=\"center\">\n") {
		t.Fatalf("header block should open the file:\n%s", got)
	}
	// The canonical block carries the title, so the original heading must be
	// replaced rather than kept above it — otherwise the README shows it twice.
	if n := strings.Count(got, "# My Tool"); n != 1 {
		t.Fatalf("title appears %d times, want 1:\n%s", n, got)
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
	// Canonical includes the language selector between tagline and badges.
	src := "<div align=\"center\">\n\n# TabelaKanban\n\n**Kanban TUI sobre markdown.**\n\n**English** · [Português](README.pt-BR.md)\n\n[![Go Version](https://img.shields.io/github/go-mod/go-version/TabelaDev/tabelakanban?style=flat-square&logo=go&logoColor=white&color=00ADD8)](go.mod)\n[![License: AGPL-3.0](https://img.shields.io/badge/license-AGPL--3.0-blue?style=flat-square)](LICENSE)\n[![Built with Bubble Tea](https://img.shields.io/badge/built%20with-Bubble%20Tea-ff69b4?style=flat-square)](https://github.com/charmbracelet/bubbletea)\n[![Powered by tabelatuiui](https://img.shields.io/badge/theme-tabelatuiui-d6b4f7?style=flat-square)](https://github.com/TabelaDev/tabelatuiui)\n\n[![ko-fi](https://ko-fi.com/img/githubbutton_sm.svg)](https://ko-fi.com/ianptkcs)\n\n</div>\n\n---\n\nBODY\n"
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

func TestUpdateHeaderPreservesPreamble(t *testing.T) {
	// Anything above the centered block — an HTML comment, an anchor — used to
	// be dropped, because the rebuilt output started at the header block and
	// never re-attached lines[:start].
	readme := "<!-- generated by tabelascaffold -->\n<a name=\"top\"></a>\n\n<div align=\"center\">\n\n# myapp\n\ntagline\n\n[![License: AGPL-3.0](https://img.shields.io/badge/x)](LICENSE)\n\n</div>\n\n---\n\n## Usage\nbody\n"
	got := updateHeader(readme, project{Name: "myapp", Org: "TabelaDev"})
	if !strings.Contains(got, "generated by tabelascaffold") {
		t.Fatalf("preamble comment lost:\n%s", got)
	}
	if !strings.Contains(got, "<a name=\"top\"></a>") {
		t.Fatalf("preamble anchor lost:\n%s", got)
	}
	if !strings.Contains(got, "## Usage") {
		t.Fatalf("body lost:\n%s", got)
	}
}

func TestUpdateHeaderIgnoresBodyBadges(t *testing.T) {
	// A badge anywhere in the body used to pass for the end of the header, so
	// every section between the header and it was replaced by the canonical
	// block. Here "## Usage" sits before a badge in "## Status".
	readme := "<div align=\"center\">\n\n# myapp\n\ntagline\n\n[![License](https://img.shields.io/badge/x)](LICENSE)\n\n</div>\n\n---\n\n## Usage\nimportant body line\n\n## Status\n[![build](https://img.shields.io/badge/build-ok)](x)\n\n## End\nlast section\n"
	got := updateHeader(readme, project{Name: "myapp", Org: "TabelaDev"})
	for _, want := range []string{"## Usage", "important body line", "## Status", "build-ok", "## End", "last section"} {
		if !strings.Contains(got, want) {
			t.Fatalf("body content %q lost:\n%s", want, got)
		}
	}
}

func TestRenderFailsOnBadTemplate(t *testing.T) {
	// An unparseable template must surface an error, not a zero-byte file.
	if _, err := render("{{ .Name", project{Name: "x"}); err == nil {
		t.Fatal("render should fail on an unterminated action")
	}
	got, err := render("bin={{.Name}}", project{Name: "myapp"})
	if err != nil {
		t.Fatalf("render on a valid template: %v", err)
	}
	if got != "bin=myapp" {
		t.Fatalf("render = %q, want %q", got, "bin=myapp")
	}
}

func TestSetupRespectsIgnore(t *testing.T) {
	// The workflows are the files setup always overwrites, so they are also the
	// ones a deliberately custom repo needs protected — tabelawebui's
	// release.yml publishes to npm and must survive a re-run.
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".github", "workflows"), 0o755); err != nil {
		t.Fatal(err)
	}
	custom := "name: Release\n# publishes to npm, not a generic GitHub release\n"
	relPath := filepath.Join(dir, ".github", "workflows", "release.yml")
	if err := os.WriteFile(relPath, []byte(custom), 0o644); err != nil {
		t.Fatal(err)
	}
	ignore := "# custom npm release\n.github/workflows/release.yml\n"
	if err := os.WriteFile(filepath.Join(dir, ignoreFile), []byte(ignore), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := setup(dir, project{Name: "tabelawebui", Org: "TabelaDev", Stack: "web"}); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(relPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != custom {
		t.Fatalf("ignored release.yml was overwritten:\n%s", got)
	}
	// The non-ignored workflow is still written.
	if _, err := os.Stat(filepath.Join(dir, ".github", "workflows", "ci.yml")); err != nil {
		t.Fatalf("ci.yml should still be written: %v", err)
	}
}

func TestDoctorSkipsIgnored(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".github", "workflows"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".github", "workflows", "release.yml"), []byte("name: totally custom\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	p := project{Name: "tabelawebui", Org: "TabelaDev", Stack: "web"}

	before, _, err := doctor(dir, p)
	if err != nil {
		t.Fatal(err)
	}
	if !hasDrift(before, filepath.Join(".github", "workflows", "release.yml")) {
		t.Fatal("custom release.yml should be reported as drift without an ignore file")
	}

	if err := os.WriteFile(filepath.Join(dir, ignoreFile), []byte(".github/workflows/release.yml\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	after, ign, err := doctor(dir, p)
	if err != nil {
		t.Fatal(err)
	}
	if hasDrift(after, filepath.Join(".github", "workflows", "release.yml")) {
		t.Fatal("ignored release.yml should not be reported as drift")
	}
	if len(ign) != 1 {
		t.Fatalf("ignore set = %v, want 1 entry", ign)
	}
}

func hasDrift(ds []drift, path string) bool {
	for _, d := range ds {
		if d.Path == path {
			return true
		}
	}
	return false
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

// A README the tool just normalized must render back byte-identical, whatever
// shape it started from. Without this, `doctor` reports drift on a repo `setup`
// had only just written.
func TestUpdateHeaderIdempotentFromAnyShape(t *testing.T) {
	p := project{Name: "meuapp", Org: "TabelaDev"}
	shapes := map[string]string{
		"bare title":     "# meuapp\n\nDescrição do projeto.\n",
		"title + body":   "# meuapp\n\nUma linha.\n\n## Uso\n\nfaz assim\n",
		"already banner": "<div align=\"center\">\n\n# meuapp\n\ntagline\n\n[![License](https://img.shields.io/badge/x)](LICENSE)\n\n</div>\n\n---\n\ncorpo\n",
		"with preamble":  "<!-- nota -->\n\n# meuapp\n\nDescrição.\n",
	}
	for name, src := range shapes {
		once := updateHeader(src, p)
		twice := updateHeader(once, p)
		if once != twice {
			t.Errorf("%s: not idempotent\n--- first ---\n%s\n--- second ---\n%s", name, once, twice)
		}
		if n := strings.Count(once, "# meuapp"); n != 1 {
			t.Errorf("%s: title appears %d times, want 1:\n%s", name, n, once)
		}
	}
}

// The bilingual pair is the whole point of the language convention: each half
// has to point at the other, and only at the other.
func TestLangSwitchPointsAtTheOtherHalf(t *testing.T) {
	en := headerBlock("MeuApp", "", project{Name: "meuapp", Org: "TabelaDev"})
	if !strings.Contains(en, "**English** · [Português](README.pt-BR.md)") {
		t.Fatalf("English header missing the selector:\n%s", en)
	}

	pt := headerBlock("MeuApp", "", project{Name: "meuapp", Org: "TabelaDev", Lang: langPtBR})
	if !strings.Contains(pt, "[English](README.md) · **Português**") {
		t.Fatalf("Portuguese header missing the selector:\n%s", pt)
	}
	if strings.Contains(pt, "README.pt-BR.md") {
		t.Fatalf("Portuguese header should not link to itself:\n%s", pt)
	}
}

// The selector sits between the tagline and the badges, which is inside the
// range updateHeader scans for custom prose. Without an explicit skip it comes
// back as "tagline" and the header accumulates a copy per run.
func TestUpdateHeaderDoesNotDuplicateLangSwitch(t *testing.T) {
	for _, p := range []project{
		{Name: "meuapp", Org: "TabelaDev"},
		{Name: "meuapp", Org: "TabelaDev", Lang: langPtBR},
	} {
		src := "# MeuApp\n\nUma linha de tagline.\n\n## Uso\n\nfaz assim\n"
		once := updateHeader(src, p)
		twice := updateHeader(once, p)
		if once != twice {
			t.Errorf("lang=%q not idempotent\n--- first ---\n%s\n--- second ---\n%s", p.Lang, once, twice)
		}
		if n := strings.Count(twice, "·"); n != 1 {
			t.Errorf("lang=%q: selector appears %d times, want 1:\n%s", p.Lang, n, twice)
		}
		if !strings.Contains(twice, "Uma linha de tagline.") {
			t.Errorf("lang=%q: tagline lost:\n%s", p.Lang, twice)
		}
	}
}

// The Portuguese CONTRIBUTING is scaffolded beside the English one, and both
// must agree on the stack-specific commands a contributor is told to run.
func TestSetupWritesBilingualContributing(t *testing.T) {
	dir := t.TempDir()
	p := project{Name: "meuapp", Org: "TabelaDev", Stack: "web"}
	if err := setup(dir, p); err != nil {
		t.Fatal(err)
	}

	en, err := os.ReadFile(filepath.Join(dir, "CONTRIBUTING.md"))
	if err != nil {
		t.Fatal(err)
	}
	pt, err := os.ReadFile(filepath.Join(dir, "CONTRIBUTING.pt-BR.md"))
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(en), "[Português](CONTRIBUTING.pt-BR.md)") {
		t.Errorf("English half does not link to the Portuguese one:\n%s", en)
	}
	if !strings.Contains(string(pt), "[English](CONTRIBUTING.md)") {
		t.Errorf("Portuguese half does not link to the English one:\n%s", pt)
	}
	// Stack-aware: a web project must not be told to run go vet.
	for name, body := range map[string]string{"en": string(en), "pt": string(pt)} {
		if !strings.Contains(body, "bun run check") {
			t.Errorf("%s: web project not told to run bun:\n%s", name, body)
		}
		if strings.Contains(body, "go vet") {
			t.Errorf("%s: web project told to run go vet:\n%s", name, body)
		}
	}
}

// doctor is how a repo learns it is missing its Portuguese half.
func TestDoctorReportsMissingBilingualHalves(t *testing.T) {
	dir := t.TempDir()
	p := project{Name: "meuapp", Org: "TabelaDev"}
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# MeuApp\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	out, _, err := doctor(dir, p)
	if err != nil {
		t.Fatal(err)
	}

	missing := map[string]bool{}
	for _, d := range out {
		if d.Reason == "faltando" {
			missing[d.Path] = true
		}
	}
	for _, want := range []string{"README.pt-BR.md", "CONTRIBUTING.pt-BR.md"} {
		if !missing[want] {
			t.Errorf("doctor did not report %s as missing: %+v", want, out)
		}
	}
}
