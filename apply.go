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
}

// setup writes the open-source scaffolding into dir (which must exist and be
// a git repo, or be an empty dir). It is idempotent: files it manages are
// overwritten with the canonical version.
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

	files := map[string]string{
		filepath.Join(workflows, "ci.yml"):   ciYAML,
		filepath.Join(issues, "bug_report.yml"):   render(bugReportYAML, p),
		filepath.Join(issues, "feature_request.yml"): render(featureRequestYAML, p),
		filepath.Join(issues, "config.yml"):  issueConfigYAML,
		filepath.Join(dir, ".github", "PULL_REQUEST_TEMPLATE.md"): prTemplateMD,
		filepath.Join(dir, "CONTRIBUTING.md"): render(contributingMD, p),
	}
	if !p.Lib {
		files[filepath.Join(workflows, "release.yml")] = render(releaseYAML, p)
	}

	// CHANGELOG and LICENSE are only created when missing, so an existing
	// project's history isn't clobbered.
	if _, err := os.Stat(filepath.Join(dir, "CHANGELOG.md")); os.IsNotExist(err) {
		files[filepath.Join(dir, "CHANGELOG.md")] = changelogMD
	}
	if _, err := os.Stat(filepath.Join(dir, "LICENSE")); os.IsNotExist(err) {
		files[filepath.Join(dir, "LICENSE")] = agplLicense
	}

	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return err
		}
	}
	return nil
}

// render fills a text/template with the project shape.
func render(tmpl string, p project) string {
	var buf bytes.Buffer
	if t, err := template.New("t").Parse(tmpl); err == nil {
		if err := t.Execute(&buf, p); err != nil {
			return tmpl
		}
	}
	return buf.String()
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
