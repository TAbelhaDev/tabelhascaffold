# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.6.0] — 2026-08-25

### Changed

- Renamed the installed binary from `tabelascaffold` to `tscaf`. Usage/help
  text, the release artifacts and the README updated to match.

## [0.5.0] — 2026-08-11

> From this entry on the changelog is written in English, per the language
> convention introduced here. Earlier entries are left as the historical record
> rather than retranslated.

### Added

- The org-wide language convention now lives in the scaffold, which is the only
  mechanism that reaches every repo without depending on anyone remembering it.
  `templates/CONTRIBUTING.md` carries the rule and is now a bilingual pair:
  English is canonical (it is what GitHub renders) and `CONTRIBUTING.pt-BR.md`
  sits beside it. `setup` writes both halves.
- README headers carry a language selector between the tagline and the badges,
  pointing at the other half of the pair. `applyBadges` normalizes both
  `README.md` and `README.pt-BR.md`. The selector goes above the badges so a
  reader who cannot read the canonical English finds the way out first, and so
  ko-fi stays adjacent to the closing `</div>`.
- `doctor` reports a missing `README.pt-BR.md` or `CONTRIBUTING.pt-BR.md`, and
  checks the header of the Portuguese README once it exists.

### Changed

- The `CONTRIBUTING` template is stack-aware: a web project is told to run
  `bun run check` and friends instead of `go vet ./...`. Both halves render from
  the same project shape, so they cannot disagree about it.

### Notes

- `setup` does not generate `README.pt-BR.md`. A translation is prose, not
  something to scaffold — `doctor` reports its absence and leaves the writing to
  a human.

## [0.4.1] — 2026-08-10

### Fixed

- `setup` num README sem badges duplicava o título: o bloco canônico já carrega
  o título, mas a linha original era mantida acima dele. O `doctor` acusava
  divergência num repo que o `setup` tinha acabado de escrever — foi assim que
  apareceu. Coberto por um teste de idempotência que roda quatro formatos de
  README (título puro, título + corpo, banner pronto, com preâmbulo).

## [0.4.0] — 2026-08-10

### Added

- `.tabelascaffoldignore`: lista de caminhos que o `setup` não sobrescreve e que
  o `doctor` não reporta como divergência. Existe porque nem toda diferença é
  dívida — o `release.yml` do `tabelawebui` publica no npm e seria substituído
  pelo workflow genérico na próxima rodada do `setup`.

## [0.3.0] — 2026-08-10

### Added

- Comando `doctor <dir>`: compara um repo com os templates canônicos e lista o
  que divergiu, sem escrever nada. Sai com código 1 quando há divergência, então
  serve de checagem em CI. É a contrapartida read-only do `setup`, que é
  destrutivo por natureza.
- O `ci.yml` dos projetos Go passou a checar `gofmt -l` — o CI validava `vet`,
  `test` e `build`, mas passava com código fora de formato.

### Fixed

- `setup` não apaga mais conteúdo do README. A varredura de badges percorria o
  arquivo inteiro, então uma badge no corpo (status de build, cobertura) passava
  por fim do cabeçalho e todo o miolo entre o header e ela era substituído pelo
  bloco canônico. Agora a busca para no `</div>` do bloco centralizado ou no
  primeiro `---`.
- `setup` preserva o que vem antes do cabeçalho (comentário HTML, âncora): o
  bloco reconstruído começava direto no header e descartava as linhas anteriores.
- Erro de parse de template agora aborta o `setup` em vez de gravar um arquivo
  vazio: `render` devolvia string vazia silenciosamente e o `ci.yml`/
  `CONTRIBUTING.md` saía com zero byte.
- `ci-web.yml` só roda `bun run test` quando o `package.json` tem esse script —
  antes o passo quebrava em todo repo web que ainda não tinha testes.
- `ci-web.yml` e `release-web.yml` ganharam newline final. Sem ela, todo repo web
  scaffoldado reprovava no próprio `prettier --check` do CI.
- `release` agora aceita `--version` depois do diretório
  (`tabelascaffold release . --version v0.2.0`): o `flag` do Go para no
  primeiro argumento posicional e o flag era ignorado silenciosamente. O
  split de args foi extraído num helper (`splitArgs`) usado pelos dois
  comandos, com teste.
