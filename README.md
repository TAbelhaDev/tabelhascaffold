<div align="center">

# tabelascaffold

**Injects the open-source structure (CI, release, templates, CONTRIBUTING,
LICENSE, CHANGELOG, badges) into a new Go Bubble Tea project.**

**English** · [Português](README.pt-BR.md)

[![Go Version](https://img.shields.io/github/go-mod/go-version/TabelaDev/tabelascaffold?style=flat-square&logo=go&logoColor=white&color=00ADD8)](go.mod)
[![License: AGPL-3.0](https://img.shields.io/badge/license-AGPL--3.0-blue?style=flat-square)](LICENSE)
[![Built with Bubble Tea](https://img.shields.io/badge/built%20with-Bubble%20Tea-ff69b4?style=flat-square)](https://github.com/charmbracelet/bubbletea)
[![Powered by tabelatuiui](https://img.shields.io/badge/theme-tabelatuiui-d6b4f7?style=flat-square)](https://github.com/TabelaDev/tabelatuiui)

[![ko-fi](https://ko-fi.com/img/githubbutton_sm.svg)](https://ko-fi.com/ianptkcs)

</div>

---

## What it is

A CLI that idempotently applies the whole open-source structure my Go TUIs
share: the CI and release workflows, the issue and PR templates, the
`CONTRIBUTING.md` pair, the AGPL-3.0 `LICENSE`, the `CHANGELOG.md` skeleton and
the README badge block (with the ko-fi button **below** the Go/License tags). A
new repo starts with the same skeleton as the mature ones.

## Installation

```bash
go install github.com/ianptkcs/tabelascaffold@latest
```

That installs the binary as `tabelascaffold` (matching the module name). To get the short
`tscaf` name used throughout this README, build from source instead:

```bash
git clone https://github.com/TabelaDev/tabelascaffold.git
cd tabelascaffold
go build -o tscaf .
```

## Usage

```bash
tscaf setup . --name myapp --org TabelaDev
tscaf setup . --name my-lib --org TabelaDev --lib   # no binary release
tscaf doctor . --name myapp --org TabelaDev         # what drifted (writes nothing)
tscaf release . --version v0.1.0                    # tag + push (workflow builds the release)
```

What `setup` creates:

```
.github/workflows/ci.yml            # vet + test + build
.github/workflows/release.yml       # linux/darwin × amd64/arm64 matrix + GitHub release
.github/ISSUE_TEMPLATE/*.yml        # bug + feature (+ config)
.github/PULL_REQUEST_TEMPLATE.md
CONTRIBUTING.md                     # English, canonical
CONTRIBUTING.pt-BR.md               # Portuguese half of the pair
LICENSE                             # AGPL-3.0 (only if absent)
CHANGELOG.md                        # Keep a Changelog skeleton (only if absent)
README.md                           # badges inserted/updated, ko-fi below the tags
README.pt-BR.md                     # header normalized when the file exists
```

`release` only creates the tag and pushes it — the release workflow builds the
binaries and publishes the GitHub release (detecting `-beta.N` prereleases).

`doctor` is the read-only counterpart of `setup`: it compares the repo against
the canonical templates, lists what drifted and exits 1 when anything did — so
it works as a CI check. It writes nothing.

### Deliberate divergence

Not every difference is debt. `tabelawebui` publishes to npm, so its
`release.yml` **cannot** be the generic one. A `.tabelascaffoldignore` at the
repo root lists the paths `setup` will not overwrite and `doctor` will not
report:

```
# own release: publishes to npm via Trusted Publishing
.github/workflows/release.yml
```

One path per line, relative to the root, `#` for comments.

## Language convention

The scaffold is where the org-wide language rule lives, because it is the only
mechanism that reaches every repo without depending on anyone remembering it.
The full rule is in [CONTRIBUTING.md](CONTRIBUTING.md#language); the short
version:

- **English, no exceptions** — everything a developer types: identifiers, file
  names, routes, query parameters, database schema, comments, commit messages.
  Brazilian domain vocabulary (`pix`, `boleto`, `cpf`) stays as-is; it is proper
  nouns, not a pending translation.
- **Bilingual** — `README.md` and `CONTRIBUTING.md`, English canonical with a
  `.pt-BR.md` half beside it and a selector at the top of each.
- **English only** — `CHANGELOG.md`, deliberately: it changes every release and
  two hand-kept copies drift.
- **Untranslated** — working notes (`AGENTS.md`, `TODO.md`, `requests/`) and
  content that *is* the product.

`setup` writes both halves of the `CONTRIBUTING` pair. It does **not** write a
`README.pt-BR.md`: a translation is prose, not something to generate. `doctor`
reports the missing half instead, and normalizes the header of both once they
exist.

## How it works

- **Embedded templates** via `go:embed` — the binary is self-contained.
- **Idempotent** — workflows are always overwritten with the canonical version;
  issue/PR templates, the `CONTRIBUTING` pair, `CHANGELOG.md` and `LICENSE` are
  only created when absent, so a project with customized templates or history is
  never clobbered.
- **Badges in the right place** — the badge block (Go Version, License, Bubble
  Tea, tabelatuiui) sits on the first line and ko-fi on the line below, always.

## Who uses it

| Project | Lib? | Origin |
|---|---|---|
| [djobs](https://github.com/ianptkcs/dankjobs) | app | already canonical (source of the pattern) |
| [tabelaradar](https://github.com/TabelaDev/tabelaradar) | app | unified |
| [tabelakanban](https://github.com/TabelaDev/tabelakanban) | app | unified |
| [tabelatuiui](https://github.com/TabelaDev/tabelatuiui) | lib | unified |

The TUI chrome itself (theme, panels, IPC helpers) comes from
[`tabelatuiui`](https://github.com/TabelaDev/tabelatuiui).

## Development

```bash
go test ./...
```

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for the version history.

## Support the project

- **Global**: [ko-fi.com/ianptkcs](https://ko-fi.com/ianptkcs)
- **Brazil (Pix)**: scan the QR below or copy the code

  <img src="pix-qr.png" alt="Pix QR" width="200" />

  <details><summary>Pix code (copy)</summary>

  ```
  00020126580014BR.GOV.BCB.PIX01365ad933b0-dcdc-4525-a736-0759902aeec65204000053039865802BR5925Ian Patrick da Costa Soar6009SAO PAULO62140510tQA85x6Dov63041FB6
  ```

  </details>

## License

[GNU AGPL-3.0](LICENSE).
