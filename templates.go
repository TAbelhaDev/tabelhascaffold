package main

import (
	_ "embed"
)

// Templates embedded so the scaffold is a single self-contained binary —
// no file next to it needed at runtime.

//go:embed templates/ci.yml
var ciYAML string

//go:embed templates/release.yml
var releaseYAML string

//go:embed templates/ci-web.yml
var ciWebYAML string

//go:embed templates/release-web.yml
var releaseWebYAML string

//go:embed templates/bug_report.yml
var bugReportYAML string

//go:embed templates/feature_request.yml
var featureRequestYAML string

//go:embed templates/config.yml
var issueConfigYAML string

//go:embed templates/PULL_REQUEST_TEMPLATE.md
var prTemplateMD string

//go:embed templates/CONTRIBUTING.md
var contributingMD string

//go:embed templates/CHANGELOG.md
var changelogMD string

//go:embed LICENSE
var agplLicense string
