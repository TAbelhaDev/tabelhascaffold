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
		// The Portuguese halves of the bilingual pair. Reported as missing, never
		// compared: a translation is prose and is allowed to diverge from the
		// template, exactly like the English half.
		"CONTRIBUTING.pt-BR.md",
		"README.pt-BR.md",
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
	// the repo is off-model. Both halves of the bilingual pair are checked; the
	// Portuguese one only when it exists, since its absence is already reported
	// above and saying it twice is noise.
	en := p
	en.Lang = ""
	if err := checkReadmeHeader(dir, "README.md", en, ign, &out, true); err != nil {
		return nil, nil, err
	}
	ptBR := p
	ptBR.Lang = langPtBR
	if err := checkReadmeHeader(dir, "README.pt-BR.md", ptBR, ign, &out, false); err != nil {
		return nil, nil, err
	}

	return out, ign, nil
}

// checkReadmeHeader appends a drift when the README variant at rel has a header
// updateHeader would rewrite. With reportMissing set, an absent file is itself
// the drift.
func checkReadmeHeader(
	dir, rel string,
	p project,
	ign ignoreSet,
	out *[]drift,
	reportMissing bool,
) error {
	if ign.has(rel) {
		return nil
	}
	data, err := os.ReadFile(filepath.Join(dir, rel))
	switch {
	case os.IsNotExist(err):
		if reportMissing {
			*out = append(*out, drift{rel, "faltando"})
		}
		return nil
	case err != nil:
		return err
	}
	if updateHeader(string(data), p) != string(data) {
		*out = append(*out, drift{rel, "cabeçalho fora do padrão"})
	}
	return nil
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
