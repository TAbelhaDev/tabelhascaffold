<div align="center">

# tabelhascaffold

**Injeta a estrutura open-source (CI, release, templates, CONTRIBUTING,
LICENSE, CHANGELOG, badges) num projeto Go Bubble Tea novo.**

[English](README.md) · **Português**

[![Go Version](https://img.shields.io/github/go-mod/go-version/TAbelhaDev/tabelhascaffold?style=flat-square&logo=go&logoColor=white&color=00ADD8)](go.mod)
[![License: AGPL-3.0](https://img.shields.io/badge/license-AGPL--3.0-blue?style=flat-square)](LICENSE)
[![Built with Bubble Tea](https://img.shields.io/badge/built%20with-Bubble%20Tea-ff69b4?style=flat-square)](https://github.com/charmbracelet/bubbletea)
[![Powered by tabelhatuiui](https://img.shields.io/badge/theme-tabelhatuiui-d6b4f7?style=flat-square)](https://github.com/TAbelhaDev/tabelhatuiui)

[![ko-fi](https://ko-fi.com/img/githubbutton_sm.svg)](https://ko-fi.com/ianptkcs)

</div>

---

## O que é

Um CLI que aplica — de forma idempotente — toda a estrutura open-source que
os meus TUIs Go compartilham: os workflows de CI e release, os templates de
issue e PR, o par `CONTRIBUTING.md`, o `LICENSE` AGPL-3.0, o `CHANGELOG.md`
skeleton e o bloco de badges do README (com o botão ko-fi **abaixo** das tags
de Go/Licença). Um repo novo começa com o mesmo esqueleto dos maduros.

## Instalação

```bash
go install github.com/TAbelhaDev/tabelhascaffold@latest
```

Isso instala o binário como `tabelhascaffold` (nome do módulo). Pra ter o nome curto
`tscaf` usado no resto deste README, compile a partir do source:

```bash
git clone https://github.com/TAbelhaDev/tabelhascaffold.git
cd tabelhascaffold
go build -o tscaf .
```

## Uso

```bash
tscaf setup . --name meuapp --org TAbelhaDev
tscaf setup . --name minha-lib --org TAbelhaDev --lib   # sem release de binário
tscaf doctor . --name meuapp --org TAbelhaDev           # o que divergiu (não escreve nada)
tscaf release . --version v0.1.0                       # tag + push (workflow gera o release)
```

O que `setup` cria:

```
.github/workflows/ci.yml            # vet + test + build
.github/workflows/release.yml       # build matriz linux/darwin × amd64/arm64 + GitHub release
.github/ISSUE_TEMPLATE/*.yml        # bug + feature (+ config)
.github/PULL_REQUEST_TEMPLATE.md
CONTRIBUTING.md                     # inglês, canônico
CONTRIBUTING.pt-BR.md               # metade em português do par
LICENSE                             # AGPL-3.0 (só se ainda não existir)
CHANGELOG.md                        # skeleton Keep a Changelog (só se não existir)
README.md                           # badges inseridos/atualizados, ko-fi abaixo das tags
README.pt-BR.md                     # cabeçalho normalizado quando o arquivo existe
```

O `release` só cria a tag e empurra — o workflow de release builda os
binários e publica o GitHub release (com detecção de prerelease `-beta.N`).

O `doctor` é a contrapartida read-only do `setup`: compara o repo com os
templates canônicos, lista o que divergiu e sai com código 1 se houver
divergência — dá pra usar como checagem em CI. Não escreve nada.

### Divergência de propósito

Nem toda diferença é dívida. O `tabelhawebui` publica no npm, então o
`release.yml` dele **não** pode ser o genérico. Um `.tabelhascaffoldignore` na
raiz do repo lista os caminhos que o `setup` não sobrescreve e que o `doctor`
não reporta:

```
# release próprio: publica no npm via Trusted Publishing
.github/workflows/release.yml
```

Um caminho por linha, relativo à raiz, `#` para comentário.

## Convenção de idioma

O scaffold é onde a regra de idioma do org mora, porque é o único mecanismo que
alcança todo repo sem depender de alguém lembrar dela. A regra completa está no
[CONTRIBUTING.pt-BR.md](CONTRIBUTING.pt-BR.md#idioma); a versão curta:

- **Inglês, sem exceção** — tudo que um dev digita: identificadores, nomes de
  arquivo, rotas, query params, schema de banco, comentários, mensagens de
  commit. Vocabulário de domínio brasileiro (`pix`, `boleto`, `cpf`) fica como
  está; é nome próprio, não tradução pendente.
- **Bilíngue** — `README.md` e `CONTRIBUTING.md`, inglês canônico com a metade
  `.pt-BR.md` ao lado e um seletor no topo de cada.
- **Só inglês** — `CHANGELOG.md`, de propósito: muda a cada release e duas
  cópias mantidas à mão dessincronizam.
- **Sem tradução** — anotação de trabalho (`AGENTS.md`, `TODO.md`, `requests/`)
  e conteúdo que *é* o produto.

O `setup` escreve as duas metades do par `CONTRIBUTING`. Ele **não** escreve
`README.pt-BR.md`: tradução é prosa, não coisa de gerar. O `doctor` reporta a
metade faltando, e normaliza o cabeçalho das duas depois que existirem.

## Como funciona

- **Templates embutidos** com `go:embed` — o binário é autocontido.
- **Idempotente** — workflows são sempre sobrescritos com a versão canônica;
  issue/PR templates, o par `CONTRIBUTING`, `CHANGELOG.md` e `LICENSE` são
  criados apenas quando não existem, então um projeto com templates
  customizados ou histórico não é sobrescrito.
- **Badges no lugar certo** — o bloco de badges (Go Version, License, Bubble
  Tea, tabelhatuiui) fica na primeira linha e o ko-fi na linha de baixo,
  sempre.

## Quem usa

| Projeto | Lib? | De onde veio |
|---|---|---|
| [djobs](https://github.com/ianptkcs/dankjobs) | app | já canônico (fonte do padrão) |
| [tabelharadar](https://github.com/TAbelhaDev/tabelharadar) | app | unificado |
| [tabelakanban](https://github.com/TAbelhaDev/tabelakanban) | app | unificado |
| [tabelhatuiui](https://github.com/TAbelhaDev/tabelhatuiui) | lib | unificado |

O chrome das TUIs em si (tema, panels, helpers de IPC) vem da
[`tabelhatuiui`](https://github.com/TAbelhaDev/tabelhatuiui).

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
