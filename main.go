package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}
	switch os.Args[1] {
	case "setup":
		os.Exit(runSetup(os.Args[2:]))
	case "release":
		os.Exit(runRelease(os.Args[2:]))
	case "help", "-h", "--help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "comando desconhecido: %q\n", os.Args[1])
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, `uso: tabelascaffold <comando> [flags]

comandos:
  setup <dir>    injeta a estrutura open-source (CI, release, templates,
                 CONTRIBUTING, LICENSE, CHANGELOG, badges do README) num repo
  release <dir>  cria a tag git e empurra, deixando o workflow de release
                 gerar os binários e o GitHub release

setup flags:
  --name X     nome do binário/módulo (padrão: basename do dir)
  --title X    título humano pro CONTRIBUTING (padrão: derivado de --name)
  --org X      org/owner do repo (padrão: TabelaDev)
  --lib        projeto biblioteca (sem workflow de release de binário)

release flags:
  --version X  versão da tag (padrão: precisa ser informada)`)
}

func runSetup(args []string) int {
	fs := flag.NewFlagSet("setup", flag.ExitOnError)
	var name, title, org string
	lib := fs.Bool("lib", false, "")
	fs.StringVar(&name, "name", "", "")
	fs.StringVar(&title, "title", "", "")
	fs.StringVar(&org, "org", "", "")

	// flag stops at the first positional arg, so split args into flags (which
	// may carry a value) and the one positional dir. Flags with a value
	// consume their following arg, so those pairs must be preserved as-is.
	dir := "."
	seenDir := false
	valueFlags := map[string]bool{"-name": true, "-title": true, "-org": true, "--name": true, "--title": true, "--org": true}
	var rest []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		if !strings.HasPrefix(a, "-") && !seenDir {
			dir = a
			seenDir = true
			continue
		}
		rest = append(rest, a)
		if valueFlags[a] && i+1 < len(args) {
			i++
			rest = append(rest, args[i])
		}
	}
	fs.Parse(rest)
	if fs.NArg() > 0 {
		dir = fs.Arg(0)
	}

	abs, err := filepath.Abs(dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		return 1
	}
	if name == "" {
		name = filepath.Base(abs)
	}
	p := project{Name: name, Title: title, Org: org, Lib: *lib}

	if err := setup(abs, p); err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		return 1
	}
	if err := applyBadges(abs, p); err != nil && !os.IsNotExist(err) {
		fmt.Fprintln(os.Stderr, "aviso (README):", err)
	}
	fmt.Printf("tabelascaffold: estrutura open-source aplicada em %s\n", abs)
	fmt.Printf("  nome=%s org=%s lib=%v\n", p.Name, p.Org, p.Lib)
	fmt.Println("  agora: tabelascaffold release . --version v0.1.0")
	return 0
}

func runRelease(args []string) int {
	fs := flag.NewFlagSet("release", flag.ExitOnError)
	version := fs.String("version", "", "")
	fs.Parse(args)
	if *version == "" {
		fmt.Fprintln(os.Stderr, "tabelascaffold: --version é obrigatório (ex: --version v0.2.0)")
		return 1
	}
	dir := "."
	if fs.NArg() > 0 {
		dir = fs.Arg(0)
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		return 1
	}
	if err := release(abs, strings.TrimPrefix(*version, "v")); err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		return 1
	}
	fmt.Printf("tabelascaffold: tag %s criada e empurrada\n", *version)
	return 0
}
