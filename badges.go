package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// badgeBlock builds the README badge block: Go version + license + bubble
// tea (plus the tabelatuiui badge for non-lib projects) on one line, and the
// ko-fi button on its own line below them — the "support" CTA always sits
// under the tech tags.
func badgeBlock(p project) string {
	goBadge := "[![Go Version](https://img.shields.io/github/go-mod/go-version/" + p.Org + "/" + p.Name + "?style=flat-square&logo=go&logoColor=white&color=00ADD8)](go.mod)"
	license := "[![License: AGPL-3.0](https://img.shields.io/badge/license-AGPL--3.0-blue?style=flat-square)](LICENSE)"
	bubble := "[![Built with Bubble Tea](https://img.shields.io/badge/built%20with-Bubble%20Tea-ff69b4?style=flat-square)](https://github.com/charmbracelet/bubbletea)"
	kofi := "[![ko-fi](https://ko-fi.com/img/githubbutton_sm.svg)](https://ko-fi.com/ianptkcs)"

	var tech []string
	tech = append(tech, goBadge, license, bubble)
	if !p.Lib {
		tech = append(tech, "[![Powered by tabelatuiui](https://img.shields.io/badge/theme-tabelatuiui-d6b4f7?style=flat-square)](https://github.com/TabelaDev/tabelatuiui)")
	}
	return strings.Join(tech, "\n") + "\n\n" + kofi
}

// updateREADMEBadges injects the badge block into the README, replacing any
// existing badge region. Returns the new content.
func updateREADMEBadges(readme, block string) string {
	badgeLine := regexp.MustCompile(`(img\.shields\.io|ko-fi\.com/img)`)

	lines := strings.Split(readme, "\n")

	// Find the badge region: the minimal span covering every badge line,
	// widened to include adjacent blank lines so the block replaces the
	// whole gap (badges + trailing blank) rather than leaving doubles.
	first, last := -1, -1
	for i, line := range lines {
		if badgeLine.MatchString(line) {
			if first == -1 {
				first = i
			}
			last = i
		}
	}

	if first == -1 {
		// No badges yet: insert after the first heading so they sit right
		// under the title.
		blockLines := strings.Split(block, "\n")
		for j, line := range lines {
			if strings.HasPrefix(line, "# ") {
				rest := append([]string{}, lines[j+1:]...)
				// Drop leading blank lines so the block's own trailing blank
				// is the single separator — matches the replace path, keeping
				// the transform idempotent.
				for len(rest) > 0 && strings.TrimSpace(rest[0]) == "" {
					rest = rest[1:]
				}
				lines = append(lines[:j+1], append(blockLines, rest...)...)
				break
			}
		}
		return strings.Join(lines, "\n")
	}

	for first > 0 && strings.TrimSpace(lines[first-1]) == "" {
		first--
	}
	for last < len(lines)-1 && strings.TrimSpace(lines[last+1]) == "" {
		last++
	}

	out := append([]string{}, lines[:first]...)
	out = append(out, block)
	out = append(out, lines[last+1:]...)
	return strings.Join(out, "\n")
}

// applyBadges updates the README of dir in place.
func applyBadges(dir string, p project) error {
	path := filepath.Join(dir, "README.md")
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	updated := updateREADMEBadges(string(data), badgeBlock(p))
	if updated == string(data) {
		return nil
	}
	return os.WriteFile(path, []byte(updated), 0o644)
}
