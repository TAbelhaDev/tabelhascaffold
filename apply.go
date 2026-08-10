package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

// project is the rendered shape of a scaffolded project.
type project struct {
	// Name is the binary/module base name (e.g. "tabelakanban").
	Name string
	// Title is the human-readable project name (e.g. "TabelaKanban").
	Title string
	// Org is the GitHub org/owner hosting the repo.
	Org string
	// Lib marks a library project: no binary release workflow.
	Lib bool
	// Stack is the project stack: "" or "tui" for Go TUIs, "web" for
	// SvelteKit/Cloudflare sites.
	Stack string
}

// web reports whether the project uses the web scaffolding.
func (p project) web() bool { return p.Stack == "web" }

// setup writes the open-source scaffolding into dir (which must exist and be
// a git repo, or be an empty dir). It is idempotent: the CI/release
// workflows are always overwritten with the canonical version; everything
// else (issue/PR templates, CONTRIBUTING, CHANGELOG, LICENSE) is only
// created when missing, so a re-run never clobbers a project's existing
// custom prose, language or history.
func setup(dir string, p project) error {
	if p.Name == "" {
		return fmt.Errorf("nome vazio")
	}
	if p.Title == "" {
		p.Title = humanizeTitle(p.Name)
	}
	if p.Org == "" {
		p.Org = "TabelaDev"
	}

	workflows := filepath.Join(dir, ".github", "workflows")
	issues := filepath.Join(dir, ".github", "ISSUE_TEMPLATE")
	for _, d := range []string{workflows, issues} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return err
		}
	}

	ciTmpl := ciYAML
	if p.web() {
		ciTmpl = ciWebYAML
	}
	files := map[string]string{
		filepath.Join(workflows, "ci.yml"): ciTmpl,
	}
	if !p.Lib {
		releaseTmpl := releaseYAML
		if p.web() {
			releaseTmpl = releaseWebYAML
		}
		releaseYML, err := render(releaseTmpl, p)
		if err != nil {
			return fmt.Errorf("release.yml: %w", err)
		}
		files[filepath.Join(workflows, "release.yml")] = releaseYML
	}

	bugReport, err := render(bugReportYAML, p)
	if err != nil {
		return fmt.Errorf("bug_report.yml: %w", err)
	}
	featureRequest, err := render(featureRequestYAML, p)
	if err != nil {
		return fmt.Errorf("feature_request.yml: %w", err)
	}
	contributing, err := render(contributingMD, p)
	if err != nil {
		return fmt.Errorf("CONTRIBUTING.md: %w", err)
	}

	// The remaining scaffold files are only created when missing, so an
	// existing project's custom prose (language, specifics) and history
	// aren't clobbered by a re-run. New projects get the canonical (PT-BR)
	// versions.
	createIfMissing := map[string]string{
		filepath.Join(issues, "bug_report.yml"):                   bugReport,
		filepath.Join(issues, "feature_request.yml"):              featureRequest,
		filepath.Join(issues, "config.yml"):                       issueConfigYAML,
		filepath.Join(dir, ".github", "PULL_REQUEST_TEMPLATE.md"): prTemplateMD,
		filepath.Join(dir, "CONTRIBUTING.md"):                     contributing,
		filepath.Join(dir, "CHANGELOG.md"):                        changelogMD,
		filepath.Join(dir, "LICENSE"):                             agplLicense,
	}
	for path, content := range createIfMissing {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			files[path] = content
		}
	}

	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return err
		}
	}
	return nil
}

// render fills a text/template with the project shape. It reports failure
// instead of falling back to something plausible: a template that failed to
// parse used to yield an empty string, and setup would happily write a
// zero-byte ci.yml or CONTRIBUTING.md with no diagnostic at all.
func render(tmpl string, p project) (string, error) {
	t, err := template.New("t").Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, p); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}
	return buf.String(), nil
}

// humanizeTitle converts a kebab/lowercase name into a title-case project
// name ("tabelakanban" -> "TabelaKanban", "djobs" -> "Djobs").
func humanizeTitle(name string) string {
	var b strings.Builder
	upper := true
	for _, r := range name {
		if r == '-' || r == '_' {
			upper = true
			continue
		}
		if upper {
			b.WriteString(strings.ToUpper(string(r)))
			upper = false
		} else {
			b.WriteString(string(r))
		}
	}
	return b.String()
}

// release tags the repo at dir with version and pushes the tag, letting the
// release workflow build the binaries and publish the GitHub release.
func release(dir, version string) error {
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}
	if err := runIn(dir, "git", "tag", version); err != nil {
		return err
	}
	if err := runIn(dir, "git", "push", "origin", version); err != nil {
		return err
	}
	return nil
}

func runIn(dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
