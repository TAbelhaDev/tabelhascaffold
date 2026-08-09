<div align="center">

# tabelascaffold

**Injeta a estrutura open-source (CI, release, templates, CONTRIBUTING,
LICENSE, CHANGELOG, badges) num projeto Go Bubble Tea novo.**

[![Go Version](https://img.shields.io/github/go-mod/go-version/TabelaDev/tabelascaffold?style=flat-square&logo=go&logoColor=white&color=00ADD8)](go.mod)
[![License: AGPL-3.0](https://img.shields.io/badge/license-AGPL--3.0-blue?style=flat-square)](LICENSE)
[![Built with Bubble Tea](https://img.shields.io/badge/built%20with-Bubble%20Tea-ff69b4?style=flat-square)](https://github.com/charmbracelet/bubbletea)
[![Powered by tabelatuiui](https://img.shields.io/badge/theme-tabelatuiui-d6b4f7?style=flat-square)](https://github.com/TabelaDev/tabelatuiui)

[![ko-fi](https://ko-fi.com/img/githubbutton_sm.svg)](https://ko-fi.com/ianptkcs)

</div>

---

## O que é

Um CLI que aplica — de forma idempotente — toda a estrutura open-source que
os meus TUIs Go compartilham: os workflows de CI e release, os templates de
issue e PR, o `CONTRIBUTING.md`, o `LICENSE` AGPL-3.0, o `CHANGELOG.md`
skeleton e o bloco de badges do README (com o botão ko-fi **abaixo** das tags
de Go/Licença). Um repo novo começa com o mesmo esqueleto dos maduros.

## Instalação

```bash
go install github.com/ianptkcs/tabelascaffold@latest
```

Ou compilando a partir do source:

```bash
git clone https://github.com/TabelaDev/tabelascaffold.git
cd tabelascaffold
go build -o tabelascaffold .
```

## Uso

```bash
tabelascaffold setup . --name meuapp --org TabelaDev
tabelascaffold setup . --name minha-lib --org TabelaDev --lib   # sem release de binário
tabelascaffold release . --version v0.1.0                       # tag + push (workflow gera o release)
```

O que `setup` cria:

```
.github/workflows/ci.yml            # vet + test + build
.github/workflows/release.yml       # build matriz linux/darwin × amd64/arm64 + GitHub release
.github/ISSUE_TEMPLATE/*.yml        # bug + feature (+ config)
.github/PULL_REQUEST_TEMPLATE.md
CONTRIBUTING.md
LICENSE                             # AGPL-3.0 (só se ainda não existir)
CHANGELOG.md                        # skeleton Keep a Changelog (só se não existir)
README.md                           # badges inseridos/atualizados, ko-fi abaixo das tags
```

O `release` só cria a tag e empurra — o workflow de release builda os
binários e publica o GitHub release (com detecção de prerelease `-beta.N`).

## Como funciona

- **Templates embutidos** com `go:embed` — o binário é autocontido.
- **Idempotente** — workflows são sempre sobrescritos com a versão canônica;
  issue/PR templates, `CONTRIBUTING.md`, `CHANGELOG.md` e `LICENSE` são
  criados apenas quando não existem, então um projeto com templates
  customizados (ex. em inglês) ou histórico não é sobrescrito.
- **Badges no lugar certo** — o bloco de badges (Go Version, License, Bubble
  Tea, tabelatuiui) fica na primeira linha e o ko-fi na linha de baixo,
  sempre.

## Quem usa

| Projeto | Lib? | De onde veio |
|---|---|---|
| [djobs](https://github.com/ianptkcs/dankjobs) | app | já canônico (fonte do padrão) |
| [tabelaradar](https://github.com/TabelaDev/tabelaradar) | app | unificado |
| [tabelakanban](https://github.com/TabelaDev/tabelakanban) | app | unificado |
| [tabelatuiui](https://github.com/TabelaDev/tabelatuiui) | lib | unificado |

O chrome das TUIs em si (tema, panels, helpers de IPC) vem da
[`tabelatuiui`](https://github.com/TabelaDev/tabelatuiui).

## Desenvolvimento

```bash
go test ./...
```

## Changelog

Veja [CHANGELOG.md](CHANGELOG.md) para o histórico de versões.

## Apoie o projeto

- **Global**: [ko-fi.com/ianptkcs](https://ko-fi.com/ianptkcs)
- **Brasil (Pix)**: escaneie o QR abaixo ou copie o código

  <img src="pix-qr.png" alt="Pix QR" width="200" />

  <details><summary>Código Pix (copiar)</summary>

  ```
  00020126580014BR.GOV.BCB.PIX01365ad933b0-dcdc-4525-a736-0759902aeec65204000053039865802BR5925Ian Patrick da Costa Soar6009SAO PAULO62140510tQA85x6Dov63041FB6
  ```

  </details>

## Licença

[GNU AGPL-3.0](LICENSE).
