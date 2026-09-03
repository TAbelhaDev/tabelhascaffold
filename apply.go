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
	// Categories are the selected scaffolding facets — any non-empty subset
	// of the registered category ids ("github", "web", "tui", and, in
	// future, "os"); see allCategories in category.go. Categories are
	// independently selectable, not mutually exclusive, and there is no
	// default — at least one must be chosen by the caller.
	Categories []string
	// Lang is which half of a bilingual doc is being rendered: "" for English —
	// the canonical one, since it is what GitHub renders — or langPtBR. Only the
	// README/CONTRIBUTING pair is bilingual; see the Language section of
	// CONTRIBUTING.md for what that covers and what it deliberately does not.
	Lang string
}

// langPtBR marks the Portuguese half of a bilingual doc, and is the suffix its
// filename carries.
const langPtBR = "pt-BR"

// hasCategory reports whether id is among p.Categories.
func (p project) hasCategory(id string) bool {
	for _, c := range p.Categories {
		if c == id {
			return true
		}
	}
	return false
}

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
		p.Org = "TAbelhaDev"
	}
	if len(p.Categories) == 0 {
		return fmt.Errorf("nenhuma categoria selecionada (%s)", strings.Join(validCategoryIDs(), ", "))
	}

	files := map[string]string{}
	createIfMissing := map[string]string{}
	for _, c := range selectedCategories(p) {
		cf, err := c.canonicalFiles(p)
		if err != nil {
			return fmt.Errorf("%s: %w", c.id(), err)
		}
		for rel, content := range cf {
			files[rel] = content
		}
		cm, err := c.createOnceFiles(p)
		if err != nil {
			return fmt.Errorf("%s: %w", c.id(), err)
		}
		for rel, content := range cm {
			createIfMissing[rel] = content
		}
	}

	// Directories are derived from the file paths the selected categories
	// declared, not hardcoded — a category that writes only to the repo root
	// (github's LICENSE, CONTRIBUTING.md) needs no subdirectory created at all.
	dirsNeeded := map[string]bool{}
	for rel := range files {
		dirsNeeded[filepath.Dir(rel)] = true
	}
	for rel := range createIfMissing {
		dirsNeeded[filepath.Dir(rel)] = true
	}
	for d := range dirsNeeded {
		if d == "." {
			continue
		}
		if err := os.MkdirAll(filepath.Join(dir, d), 0o755); err != nil {
			return err
		}
	}

	// The remaining scaffold files are only created when missing, so an
	// existing project's custom prose (language, specifics) and history
	// aren't clobbered by a re-run.
	for rel, content := range createIfMissing {
		if _, err := os.Stat(filepath.Join(dir, rel)); os.IsNotExist(err) {
			files[rel] = content
		}
	}

	// A repo may deliberately diverge from the canonical scaffold (tabelhawebui's
	// release.yml publishes to npm). Those paths are declared in
	// .tabelhascaffoldignore and setup leaves them alone.
	ign, err := loadIgnore(dir)
	if err != nil {
		return fmt.Errorf("%s: %w", ignoreFile, err)
	}

	for rel, content := range files {
		if ign.has(rel) {
			continue
		}
		if err := os.WriteFile(filepath.Join(dir, rel), []byte(content), 0o644); err != nil {
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
