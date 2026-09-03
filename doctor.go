package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
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
		p.Org = "TAbelhaDev"
	}
	if len(p.Categories) == 0 {
		return nil, nil, fmt.Errorf("nenhuma categoria selecionada (%s)", strings.Join(validCategoryIDs(), ", "))
	}

	// Paths the repo declared as deliberately custom are neither compared nor
	// reported — otherwise doctor stays red forever on a divergence that is the
	// whole point (tabelhawebui's npm release workflow).
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

	for _, c := range selectedCategories(p) {
		cf, err := c.canonicalFiles(p)
		if err != nil {
			return nil, nil, fmt.Errorf("%s: %w", c.id(), err)
		}
		for _, rel := range sortedKeys(cf) {
			check(rel, cf[rel])
		}

		// These are create-if-missing in setup, so drift here means "absent",
		// not "different" — a project's own prose is allowed to diverge.
		cm, err := c.createOnceFiles(p)
		if err != nil {
			return nil, nil, fmt.Errorf("%s: %w", c.id(), err)
		}
		for _, rel := range sortedKeys(cm) {
			if ign.has(rel) {
				continue
			}
			if _, err := os.Stat(filepath.Join(dir, rel)); os.IsNotExist(err) {
				out = append(out, drift{rel, "faltando"})
			}
		}
	}

	// README.pt-BR.md is create-if-missing conceptually (a translation is
	// prose, github doesn't scaffold it) but isn't one of github's
	// createOnceFiles since tabelhascaffold never creates it — only
	// checkReadmeHeader below reports its absence when github is selected.
	if p.hasCategory("github") && !ign.has("README.pt-BR.md") {
		if _, err := os.Stat(filepath.Join(dir, "README.pt-BR.md")); os.IsNotExist(err) {
			out = append(out, drift{"README.pt-BR.md", "faltando"})
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

// sortedKeys returns m's keys sorted, so drift reporting is deterministic
// (map iteration order is not) within one category's file set.
func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
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
