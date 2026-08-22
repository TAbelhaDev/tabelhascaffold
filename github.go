package main

import (
	_ "embed"
	"fmt"
	"path/filepath"
)

//go:embed templates/github/bug_report.yml
var bugReportYAML string

//go:embed templates/github/feature_request.yml
var featureRequestYAML string

//go:embed templates/github/config.yml
var issueConfigYAML string

//go:embed templates/github/PULL_REQUEST_TEMPLATE.md
var prTemplateMD string

//go:embed templates/github/CONTRIBUTING.md
var contributingMD string

//go:embed templates/github/CONTRIBUTING.pt-BR.md
var contributingPtBrMD string

//go:embed templates/github/CHANGELOG.md
var changelogMD string

//go:embed LICENSE
var agplLicense string

// githubCategory is the "open source project structure" facet: LICENSE,
// CHANGELOG.md, the bilingual CONTRIBUTING pair, GitHub issue/PR templates,
// and the OSS-specific README badges (AGPL-3.0 license, ko-fi). Selectable
// independently of web/tui so a closed-source repo can take CI structure
// without this metadata.
type githubCategory struct{}

func (githubCategory) id() string { return "github" }

func (githubCategory) canonicalFiles(p project) (map[string]string, error) {
	return map[string]string{}, nil
}

func (githubCategory) createOnceFiles(p project) (map[string]string, error) {
	bugReport, err := render(bugReportYAML, p)
	if err != nil {
		return nil, fmt.Errorf("bug_report.yml: %w", err)
	}
	featureRequest, err := render(featureRequestYAML, p)
	if err != nil {
		return nil, fmt.Errorf("feature_request.yml: %w", err)
	}
	contributing, err := render(contributingMD, p)
	if err != nil {
		return nil, fmt.Errorf("CONTRIBUTING.md: %w", err)
	}
	// The Portuguese half is rendered from the same project shape, only with
	// the language marker flipped, so the two halves cannot disagree about
	// the build/test commands they tell a contributor to run.
	ptBR := p
	ptBR.Lang = langPtBR
	contributingPtBr, err := render(contributingPtBrMD, ptBR)
	if err != nil {
		return nil, fmt.Errorf("CONTRIBUTING.pt-BR.md: %w", err)
	}
	return map[string]string{
		filepath.Join(".github", "ISSUE_TEMPLATE", "bug_report.yml"):      bugReport,
		filepath.Join(".github", "ISSUE_TEMPLATE", "feature_request.yml"): featureRequest,
		filepath.Join(".github", "ISSUE_TEMPLATE", "config.yml"):          issueConfigYAML,
		filepath.Join(".github", "PULL_REQUEST_TEMPLATE.md"):              prTemplateMD,
		"CONTRIBUTING.md":       contributing,
		"CONTRIBUTING.pt-BR.md": contributingPtBr,
		"CHANGELOG.md":          changelogMD,
		"LICENSE":               agplLicense,
	}, nil
}

func (githubCategory) badges(p project) []string {
	return []string{
		"[![License: AGPL-3.0](https://img.shields.io/badge/license-AGPL--3.0-blue?style=flat-square)](LICENSE)",
	}
}

func (githubCategory) footer(p project) string { return kofi }

func (githubCategory) buildTestCommands(p project) string { return "" }
func (githubCategory) buildTestLabel(p project) string    { return "" }
