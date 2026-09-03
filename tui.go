package main

import (
	_ "embed"
	"fmt"
	"path/filepath"
)

//go:embed templates/tui/ci.yml
var ciYAML string

//go:embed templates/tui/release.yml
var releaseYAML string

// tuiCategory scaffolds a Go/Bubble Tea TUI project: the Go CI workflow, the
// binary release workflow (unless the project is a library, independent of
// any other selected category), and the Go-version/Bubble-Tea README badges.
type tuiCategory struct{}

func (tuiCategory) id() string { return "tui" }

func (tuiCategory) canonicalFiles(p project) (map[string]string, error) {
	files := map[string]string{
		filepath.Join(".github", "workflows", "ci.yml"): ciYAML,
	}
	if !p.Lib {
		rel, err := render(releaseYAML, p)
		if err != nil {
			return nil, fmt.Errorf("release.yml: %w", err)
		}
		files[filepath.Join(".github", "workflows", "release.yml")] = rel
	}
	return files, nil
}

func (tuiCategory) createOnceFiles(p project) (map[string]string, error) {
	return map[string]string{}, nil
}

func (tuiCategory) badges(p project) []string {
	b := []string{
		"[![Go Version](https://img.shields.io/github/go-mod/go-version/" + p.Org + "/" + p.Name + "?style=flat-square&logo=go&logoColor=white&color=00ADD8)](go.mod)",
		"[![Built with Bubble Tea](https://img.shields.io/badge/built%20with-Bubble%20Tea-ff69b4?style=flat-square)](https://github.com/charmbracelet/bubbletea)",
	}
	if !p.Lib {
		b = append(b, "[![Powered by tabelhatuiui](https://img.shields.io/badge/theme-tabelhatuiui-d6b4f7?style=flat-square)](https://github.com/TAbelhaDev/tabelhatuiui)")
	}
	return b
}

func (tuiCategory) footer(p project) string { return "" }

func (tuiCategory) buildTestCommands(p project) string {
	return "`go vet ./...`, `go test ./...` and `go build ./...`"
}
func (tuiCategory) buildTestLabel(p project) string { return "Go" }
