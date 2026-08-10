package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// drift is one divergence between a repo and the canonical scaffold.
type drift struct {
	Path   string
	Reason string
}

// doctor is the read-only counterpart of setup: it renders the same canonical
// files for the same project shape, compares them against what the repo has,
// and reports the differences without writing anything. It exists because
// setup is destructive by nature — this is the way to ask "how far has this
// repo drifted?" without committing to overwriting it.
func doctor(dir string, p project) ([]drift, ignoreSet, error) {
	if p.Name == "" {
		return nil, nil, fmt.Errorf("nome vazio")
	}
	if p.Title == "" {
		p.Title = humanizeTitle(p.Name)
	}
	if p.Org == "" {
		p.Org = "TabelaDev"
	}

	// Paths the repo declared as deliberately custom are neither compared nor
	// reported — otherwise doctor stays red forever on a divergence that is the
	// whole point (tabelawebui's npm release workflow).
	ign, err := loadIgnore(dir)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", ignoreFile, err)
	}

	var out []drift
	check := func(rel, want string) {
		if ign.has(rel) {
			return
		}
		out = append(out, compareFile(dir, rel, want)...)
	}

	ciTmpl := ciYAML
	if p.web() {
		ciTmpl = ciWebYAML
	}
	check(filepath.Join(".github", "workflows", "ci.yml"), ciTmpl)

	if !p.Lib {
		releaseTmpl := releaseYAML
		if p.web() {
			releaseTmpl = releaseWebYAML
		}
		want, err := render(releaseTmpl, p)
		if err != nil {
			return nil, nil, fmt.Errorf("release.yml: %w", err)
		}
		check(filepath.Join(".github", "workflows", "release.yml"), want)
	}

	// These are create-if-missing in setup, so drift here means "absent", not
	// "different" — a project's own prose is allowed to diverge.
	for _, rel := range []string{
		filepath.Join(".github", "ISSUE_TEMPLATE", "bug_report.yml"),
		filepath.Join(".github", "ISSUE_TEMPLATE", "feature_request.yml"),
		filepath.Join(".github", "ISSUE_TEMPLATE", "config.yml"),
		filepath.Join(".github", "PULL_REQUEST_TEMPLATE.md"),
		"CONTRIBUTING.md",
		"CHANGELOG.md",
		"LICENSE",
	} {
		if ign.has(rel) {
			continue
		}
		if _, err := os.Stat(filepath.Join(dir, rel)); os.IsNotExist(err) {
			out = append(out, drift{rel, "faltando"})
		}
	}

	// The README header is canonical too: if updateHeader would rewrite it,
	// the repo is off-model.
	if !ign.has("README.md") {
		data, err := os.ReadFile(filepath.Join(dir, "README.md"))
		switch {
		case os.IsNotExist(err):
			out = append(out, drift{"README.md", "faltando"})
		case err != nil:
			return nil, nil, err
		default:
			if updateHeader(string(data), p) != string(data) {
				out = append(out, drift{"README.md", "cabeçalho fora do padrão"})
			}
		}
	}

	return out, ign, nil
}

// compareFile reports whether the repo's copy of rel matches the canonical
// content byte for byte.
func compareFile(dir, rel, want string) []drift {
	got, err := os.ReadFile(filepath.Join(dir, rel))
	switch {
	case os.IsNotExist(err):
		return []drift{{rel, "faltando"}}
	case err != nil:
		return []drift{{rel, "ilegível: " + err.Error()}}
	case string(got) != want:
		return []drift{{rel, "difere do template canônico"}}
	}
	return nil
}
