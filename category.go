package main

import "strings"

// category is one independently selectable scaffolding facet — github, web,
// tui, and (future) os. setup and doctor drive the registry of selected
// categories generically instead of branching on a "stack" string.
type category interface {
	// id is the flag name and the string stored in project.Categories.
	id() string

	// canonicalFiles returns files this category always overwrites, keyed by
	// path relative to the repo root. p has Name/Title/Org already defaulted.
	canonicalFiles(p project) (map[string]string, error)

	// createOnceFiles returns files this category writes only when the repo
	// doesn't already have them, keyed by path relative to the repo root.
	createOnceFiles(p project) (map[string]string, error)

	// badges returns this category's README badge lines, in this category's
	// own preferred order. Categories are visited in registry order (see
	// allCategories), which is what fixes cross-category badge order (stack
	// badges before github's license badge).
	badges(p project) []string

	// footer returns a single badge rendered as its own paragraph below the
	// main badge block (github's ko-fi button), or "" if this category has
	// none.
	footer(p project) string

	// buildTestCommands returns the backtick-quoted command list a
	// contributor should run for this category, or "" if this category
	// prescribes none (github has none — it is metadata, not a stack).
	buildTestCommands(p project) string

	// buildTestLabel names this category's commands when more than one
	// category contributes to CONTRIBUTING's build/test step (e.g. "Go",
	// "Web"). Unused when buildTestCommands is "" or is the sole contributor.
	buildTestLabel(p project) string
}

// allCategories is the registry of every category tabelascaffold knows, in
// the fixed order that decides badge and CONTRIBUTING precedence when more
// than one is selected (stack categories first, github's license+ko-fi
// last). Adding an "os" category later is one new file (os.go) implementing
// category, plus one line here, plus one bool flag in main.go — nothing else
// changes.
var allCategories = []category{
	tuiCategory{},
	webCategory{},
	pyscriptCategory{},
	githubCategory{},
}

// validCategoryIDs lists every registered id, for usage/error text.
func validCategoryIDs() []string {
	ids := make([]string, len(allCategories))
	for i, c := range allCategories {
		ids[i] = c.id()
	}
	return ids
}

// selectedCategories returns the registered categories p.Categories names, in
// registry order — regardless of the order they appear in p.Categories.
func selectedCategories(p project) []category {
	var out []category
	for _, c := range allCategories {
		if p.hasCategory(c.id()) {
			out = append(out, c)
		}
	}
	return out
}

// BuildTestStep composes CONTRIBUTING's step-3 sentence from the stack
// categories selected on p:
//   - zero contributing categories (github alone, or no stack categories):
//     a generic placeholder — tabelascaffold has no commands to prescribe.
//   - exactly one: "Run <commands> locally before opening the PR."
//   - two or more: an intro line plus one bullet per contributing category,
//     in registry order, so a multi-stack repo (e.g. web+tui) gets both sets
//     of commands instead of only the first one silently winning.
//
// p.Lang selects English or Portuguese prose; the commands themselves (shell
// invocations) are language-neutral. Exported so text/template's
// {{.BuildTestStep}} can call it as a niladic method.
func (p project) BuildTestStep() string {
	type contributor struct{ label, commands string }
	var cs []contributor
	for _, c := range selectedCategories(p) {
		if cmds := c.buildTestCommands(p); cmds != "" {
			cs = append(cs, contributor{c.buildTestLabel(p), cmds})
		}
	}
	pt := p.Lang == langPtBR
	switch len(cs) {
	case 0:
		if pt {
			return "Rode os comandos de build/test do seu projeto aqui antes de abrir o PR."
		}
		return "Run your project's build/test commands here before opening the PR."
	case 1:
		if pt {
			return "Rode " + cs[0].commands + " localmente antes de abrir o PR."
		}
		return "Run " + cs[0].commands + " locally before opening the PR."
	default:
		intro, verb := "Depending on what you're changing:", "run"
		if pt {
			intro, verb = "Dependendo do que você está mudando:", "rode"
		}
		var b strings.Builder
		b.WriteString(intro + "\n")
		for _, c := range cs {
			b.WriteString("   - " + c.label + ": " + verb + " " + c.commands + ".\n")
		}
		return strings.TrimRight(b.String(), "\n")
	}
}
