package main

import (
	_ "embed"
	"fmt"
	"path/filepath"
)

//go:embed templates/pyscript/ci.yml
var pyscriptCiYAML string

//go:embed templates/pyscript/release.yml
var pyscriptReleaseYAML string

//go:embed templates/pyscript/pyproject.toml
var pyscriptPyproject string

// pyscriptCategory scaffolds a Python (uv) project: the uv CI workflow, the
// release workflow (tag -> GitHub release; script tools ship no binary), and the
// modern Python toolchain badges. The pyproject.toml skeleton (uv-managed: ruff
// + basedpyright + ty + typos + vulture + pytest + hypothesis + pip-audit) is
// written only when the repo doesn't already have one.
type pyscriptCategory struct{}

func (pyscriptCategory) id() string { return "pyscript" }

func (pyscriptCategory) canonicalFiles(p project) (map[string]string, error) {
	files := map[string]string{
		filepath.Join(".github", "workflows", "ci.yml"): pyscriptCiYAML,
	}
	if !p.Lib {
		rel, err := render(pyscriptReleaseYAML, p)
		if err != nil {
			return nil, fmt.Errorf("release.yml: %w", err)
		}
		files[filepath.Join(".github", "workflows", "release.yml")] = rel
	}
	return files, nil
}

func (pyscriptCategory) createOnceFiles(p project) (map[string]string, error) {
	pp, err := render(pyscriptPyproject, p)
	if err != nil {
		return nil, fmt.Errorf("pyproject.toml: %w", err)
	}
	return map[string]string{"pyproject.toml": pp}, nil
}

func (pyscriptCategory) badges(p project) []string {
	b := []string{
		"[![Python](https://img.shields.io/badge/python-3.12+-3776AB?style=flat-square&logo=python&logoColor=white)](pyproject.toml)",
		"[![uv](https://img.shields.io/badge/uv-astro-DEA584?style=flat-square&logo=astral&logoColor=white)](https://github.com/astral-sh/uv)",
		"[![typos](https://img.shields.io/badge/typos-checked-1B1FCA?style=flat-square)](https://github.com/astral-sh/typos)",
	}
	return b
}

func (pyscriptCategory) footer(p project) string { return "" }

func (pyscriptCategory) buildTestCommands(p project) string {
	return "`uv sync`, `uv run ruff check .`, `uv run basedpyright` and `uv run pytest`"
}
func (pyscriptCategory) buildTestLabel(p project) string { return "Python" }
