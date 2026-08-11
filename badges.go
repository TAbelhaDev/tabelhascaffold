package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	badgeLine = regexp.MustCompile(`(img\.shields\.io|ko-fi\.com/img)`)
	divider   = regexp.MustCompile(`^\s*-{3,}\s*$`)
	divOpen   = regexp.MustCompile(`^\s*<div`)
	divClose  = regexp.MustCompile(`^\s*<\/div`)
	// The language selector sits in the header between the tagline and the
	// badges, so the tagline scan has to recognise and skip it — otherwise a
	// re-run would take it for custom prose and the README would end up with two
	// copies of the line.
	langLine = regexp.MustCompile(`\[(English|Português)\]\(README(\.pt-BR)?\.md\)`)
)

// techBadges returns the badge lines proper for the project stack — license
// + tech tags for the header; ko-fi is handled separately by headerBlock.
func techBadges(p project) []string {
	switch {
	case p.web():
		b := []string{
			"[![SvelteKit](https://img.shields.io/badge/SvelteKit-Svelte-ff3e00?style=flat-square&logo=svelte&logoColor=white)](https://kit.svelte.dev)",
			"[![Cloudflare Workers](https://img.shields.io/badge/Cloudflare-Workers-orange?style=flat-square&logo=cloudflare&logoColor=white)](https://workers.cloudflare.com)",
			"[![License: AGPL-3.0](https://img.shields.io/badge/license-AGPL--3.0-blue?style=flat-square)](LICENSE)",
		}
		if !p.Lib {
			b = append(b, "[![Built with tabelawebui](https://img.shields.io/badge/theme-tabelawebui-d6b4f7?style=flat-square)](https://github.com/TabelaDev/tabelawebui)")
		}
		return b
	default:
		b := []string{
			"[![Go Version](https://img.shields.io/github/go-mod/go-version/" + p.Org + "/" + p.Name + "?style=flat-square&logo=go&logoColor=white&color=00ADD8)](go.mod)",
			"[![License: AGPL-3.0](https://img.shields.io/badge/license-AGPL--3.0-blue?style=flat-square)](LICENSE)",
			"[![Built with Bubble Tea](https://img.shields.io/badge/built%20with-Bubble%20Tea-ff69b4?style=flat-square)](https://github.com/charmbracelet/bubbletea)",
		}
		if !p.Lib {
			b = append(b, "[![Powered by tabelatuiui](https://img.shields.io/badge/theme-tabelatuiui-d6b4f7?style=flat-square)](https://github.com/TabelaDev/tabelatuiui)")
		}
		return b
	}
}

const kofi = "[![ko-fi](https://ko-fi.com/img/githubbutton_sm.svg)](https://ko-fi.com/ianptkcs)"

// langSwitch is the selector between the two halves of the bilingual README.
// It goes above the badges rather than below the ko-fi button: a reader who
// cannot read the canonical English should find the way out before anything
// else, and ko-fi stays adjacent to the closing </div> where it belongs.
func langSwitch(p project) string {
	if p.Lang == langPtBR {
		return "[English](README.md) · **Português**"
	}
	return "**English** · [Português](README.pt-BR.md)"
}

// headerBlock builds the canonical README header — a centered <div> holding
// the project title, its tagline, the tech badges, and the ko-fi button
// separated on its own line, then the --- divider that closes the header.
func headerBlock(title, tagline string, p project) string {
	var b strings.Builder
	b.WriteString(`<div align="center">`)
	b.WriteString("\n\n# " + title + "\n")
	if strings.TrimSpace(tagline) != "" {
		b.WriteString("\n" + tagline + "\n")
	}
	b.WriteString("\n" + langSwitch(p) + "\n")
	b.WriteString("\n" + strings.Join(techBadges(p), "\n"))
	b.WriteString("\n\n" + kofi)
	b.WriteString("\n\n</div>")
	return b.String()
}

// updateHeader rewrites the README top block to the canonical model,
// idempotently. It finds the header region — the title through the last
// badge/ko-fi line, extended to a closing </div> when one sits after it —
// and rebuilds it as <div align="center"> + preserved title + preserved
// tagline + canonical badges + ko-fi on its own line + a --- divider before
// the untouched body.
//
// Projects that already follow the model render back to byte-identical
// output on a re-run. Older bare headers (title, badges, no wrapper) get the
// <div align="center"> wrapper introduced.
func updateHeader(readme string, p project) string {
	lines := strings.Split(readme, "\n")

	titleIdx := -1
	for i, l := range lines {
		if strings.TrimSpace(l) == "" {
			continue
		}
		if strings.HasPrefix(strings.TrimSpace(l), "# ") {
			titleIdx = i
			break
		}
	}

	// The badge scan is bounded to the header region. Scanning the whole file
	// let a badge in the body (build status, coverage, a sub-package version)
	// pass for the end of the header, so everything between the header and it
	// got replaced by the canonical block — silent content loss.
	limit := headerLimit(lines)

	firstBadge, lastBadge := -1, -1
	for i := 0; i <= limit; i++ {
		if badgeLine.MatchString(lines[i]) {
			if firstBadge < 0 {
				firstBadge = i
			}
			lastBadge = i
		}
	}

	// The tagline is whatever prose sits between the title and the first
	// badge — kept verbatim so custom copy survives a re-run.
	tagline := ""
	if titleIdx >= 0 && firstBadge > titleIdx+1 {
		var prose []string
		for _, l := range lines[titleIdx+1 : firstBadge] {
			t := strings.TrimSpace(l)
			if t == "" {
				if len(prose) > 0 {
					prose = append(prose, "")
				}
				continue
			}
			if badgeLine.MatchString(l) || divClose.MatchString(l) || divOpen.MatchString(l) {
				continue
			}
			if langLine.MatchString(l) {
				continue
			}
			prose = append(prose, l)
		}
		tagline = strings.Trim(strings.Join(prose, "\n"), "\n")
	}

	// Determine the replaced span: title (or earliest badge) through the last
	// badge line, widened to an existing </div> right after the badges.
	var out []string
	title := ""
	if titleIdx >= 0 {
		title = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(lines[titleIdx]), "# "))
	}
	if title == "" {
		title = p.Title
		if title == "" {
			title = humanizeTitle(p.Name)
		}
	}
	block := strings.Split(headerBlock(title, tagline, p), "\n")

	if lastBadge < 0 {
		// No badges yet: the canonical block replaces the title line, since the
		// block already carries the title. Keeping lines[:titleIdx+1] here left
		// the original heading above the block and the README ended up with the
		// title twice.
		if titleIdx < 0 {
			return readme
		}
		out = append(out, lines[:titleIdx]...)
		out = append(out, block...)
		body := lines[titleIdx+1:]
		for len(body) > 0 && strings.TrimSpace(body[0]) == "" {
			body = body[1:]
		}
		if len(body) > 0 {
			out = append(out, "", "---", "")
			out = append(out, body...)
		}
		return strings.Join(out, "\n")
	}

	start := firstBadge
	if titleIdx >= 0 && titleIdx < firstBadge {
		start = titleIdx
	}
	for i := start - 1; i >= 0; i-- {
		if divOpen.MatchString(lines[i]) {
			start = i
			break
		}
	}
	end := lastBadge
	for i := lastBadge + 1; i < len(lines); i++ {
		if divClose.MatchString(lines[i]) {
			end = i
			break
		}
		if divider.MatchString(lines[i]) {
			break
		}
	}
	rest := lines[end+1:]
	for len(rest) > 0 && strings.TrimSpace(rest[0]) == "" {
		rest = rest[1:]
	}
	if len(rest) > 0 && divider.MatchString(rest[0]) {
		rest = rest[1:]
	}
	for len(rest) > 0 && strings.TrimSpace(rest[0]) == "" {
		rest = rest[1:]
	}

	// Whatever sat above the header (HTML comment, anchor, a "generated by"
	// notice) is preserved — only the header block itself is rewritten. Same
	// care the no-badges branch already took with lines[:titleIdx+1].
	out = append(out, lines[:start]...)
	out = append(out, block...)
	if len(rest) > 0 {
		out = append(out, "", "---", "")
		out = append(out, rest...)
	}
	return strings.Join(out, "\n")
}

// headerLimit returns the index of the last line still belonging to the
// header: the </div> that closes the centered block, or the first --- divider
// when there is no block. Badges past that point belong to the README body and
// must not be mistaken for header badges.
func headerLimit(lines []string) int {
	for i, l := range lines {
		if divClose.MatchString(l) || divider.MatchString(l) {
			return i
		}
	}
	return len(lines) - 1
}

// applyBadges normalizes the README header of dir in place, unless the repo
// opted the README out in .tabelascaffoldignore.
//
// Both halves of the bilingual pair are normalized: README.md always, and
// README.pt-BR.md when it exists. The Portuguese half is not created here — a
// translation is prose, not something to scaffold — so a repo without one is
// left alone and reported by doctor instead.
func applyBadges(dir string, p project) error {
	ign, err := loadIgnore(dir)
	if err != nil {
		return err
	}

	en := p
	en.Lang = ""
	if err := normalizeReadme(dir, "README.md", en, ign, true); err != nil {
		return err
	}

	ptBR := p
	ptBR.Lang = langPtBR
	return normalizeReadme(dir, "README.pt-BR.md", ptBR, ign, false)
}

// normalizeReadme rewrites one README variant's header. With required set, a
// missing file is an error the caller surfaces; without it, absence is fine.
func normalizeReadme(dir, rel string, p project, ign ignoreSet, required bool) error {
	if ign.has(rel) {
		return nil
	}

	path := filepath.Join(dir, rel)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) && !required {
		return nil
	}
	if err != nil {
		return err
	}
	updated := updateHeader(string(data), p)
	if updated == string(data) {
		return nil
	}
	return os.WriteFile(path, []byte(updated), 0o644)
}
