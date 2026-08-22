package main

import (
	_ "embed"
	"path/filepath"
)

//go:embed templates/web/ci-web.yml
var ciWebYAML string

// webCategory scaffolds a SvelteKit/Cloudflare Workers project: the Bun CI
// workflow (which already includes deploy + tagged-release jobs — web ships
// via Cloudflare, not a separate GoReleaser-style release.yml) and the
// SvelteKit/Cloudflare README badges.
type webCategory struct{}

func (webCategory) id() string { return "web" }

func (webCategory) canonicalFiles(p project) (map[string]string, error) {
	return map[string]string{
		filepath.Join(".github", "workflows", "ci-web.yml"): ciWebYAML,
	}, nil
}

func (webCategory) createOnceFiles(p project) (map[string]string, error) {
	return map[string]string{}, nil
}

func (webCategory) badges(p project) []string {
	b := []string{
		"[![SvelteKit](https://img.shields.io/badge/SvelteKit-Svelte-ff3e00?style=flat-square&logo=svelte&logoColor=white)](https://kit.svelte.dev)",
		"[![Cloudflare Workers](https://img.shields.io/badge/Cloudflare-Workers-orange?style=flat-square&logo=cloudflare&logoColor=white)](https://workers.cloudflare.com)",
	}
	if !p.Lib {
		b = append(b, "[![Built with tabelawebui](https://img.shields.io/badge/theme-tabelawebui-d6b4f7?style=flat-square)](https://github.com/TabelaDev/tabelawebui)")
	}
	return b
}

func (webCategory) footer(p project) string { return "" }

func (webCategory) buildTestCommands(p project) string {
	return "`bun run check`, `bun run lint`, `bun run test` and `bun run build`"
}
func (webCategory) buildTestLabel(p project) string { return "Web" }
