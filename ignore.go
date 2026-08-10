package main

import (
	"bufio"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ignoreFile is the per-repo opt-out list, one repo-relative path per line,
// `#` for comments.
//
// setup rewrites the CI and release workflows unconditionally, which is right
// for a repo that follows the canonical scaffold and wrong for one that
// deliberately diverges: tabelawebui's release.yml publishes to npm and would
// be replaced by the generic notes-only workflow on the next run. Listing it
// here keeps setup off it and stops doctor from reporting it as drift forever.
const ignoreFile = ".tabelascaffoldignore"

// ignoreSet holds the repo-relative paths setup must not write and doctor must
// not flag.
type ignoreSet map[string]bool

// loadIgnore reads the opt-out list from dir. A missing file is not an error —
// it just means nothing is exempt.
func loadIgnore(dir string) (ignoreSet, error) {
	f, err := os.Open(filepath.Join(dir, ignoreFile))
	if os.IsNotExist(err) {
		return ignoreSet{}, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()

	out := ignoreSet{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		out[filepath.Clean(line)] = true
	}
	return out, sc.Err()
}

// has reports whether a repo-relative path is exempt.
func (s ignoreSet) has(rel string) bool {
	return s[filepath.Clean(rel)]
}

// hasAbs is has() for a path under dir, for callers holding absolute paths.
func (s ignoreSet) hasAbs(dir, path string) bool {
	rel, err := filepath.Rel(dir, path)
	if err != nil {
		return false
	}
	return s.has(rel)
}

// sorted returns the exempt paths in a stable order, for reporting.
func (s ignoreSet) sorted() []string {
	out := make([]string, 0, len(s))
	for p := range s {
		out = append(out, p)
	}
	sort.Strings(out)
	return out
}
