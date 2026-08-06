# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed

- `release` agora aceita `--version` depois do diretório
  (`tabelascaffold release . --version v0.2.0`): o `flag` do Go para no
  primeiro argumento posicional e o flag era ignorado silenciosamente. O
  split de args foi extraído num helper (`splitArgs`) usado pelos dois
  comandos, com teste.
