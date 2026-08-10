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
	case "doctor":
		os.Exit(runDoctor(os.Args[2:]))
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
  doctor <dir>   compara o repo com os templates canônicos e lista o que
                 divergiu — não escreve nada; sai com código 1 se houver
                 divergência
  release <dir>  cria a tag git e empurra, deixando o workflow de release
                 gerar os binários e o GitHub release

setup/doctor flags:
  --name X     nome do binário/módulo (padrão: basename do dir)
  --title X    título humano pro CONTRIBUTING (padrão: derivado de --name)
  --org X      org/owner do repo (padrão: TabelaDev)
  --lib        projeto biblioteca (sem workflow de release de binário)
  --stack X    stack do projeto: tui (Go/Bubble Tea, padrão) ou web
               (SvelteKit/Cloudflare; CI Bun + release sem binário)

release flags:
  --version X  versão da tag (padrão: precisa ser informada)`)
}

func runSetup(args []string) int {
	fs := flag.NewFlagSet("setup", flag.ExitOnError)
	var name, title, org, stack string
	lib := fs.Bool("lib", false, "")
	fs.StringVar(&name, "name", "", "")
	fs.StringVar(&title, "title", "", "")
	fs.StringVar(&org, "org", "", "")
	fs.StringVar(&stack, "stack", "tui", "")

	dir, rest := splitArgs(args, map[string]bool{
		"-name": true, "--name": true,
		"-title": true, "--title": true,
		"-org": true, "--org": true,
		"-stack": true, "--stack": true,
	})
	fs.Parse(rest)
	if fs.NArg() > 0 {
		dir = fs.Arg(0)
	}

	if stack != "tui" && stack != "web" {
		fmt.Fprintf(os.Stderr, "erro: stack desconhecido %q (esperado tui ou web)\n", stack)
		return 1
	}

	abs, err := filepath.Abs(dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		return 1
	}
	if name == "" {
		name = filepath.Base(abs)
	}
	p := project{Name: name, Title: title, Org: org, Lib: *lib, Stack: stack}

	if err := setup(abs, p); err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		return 1
	}
	if err := applyBadges(abs, p); err != nil && !os.IsNotExist(err) {
		fmt.Fprintln(os.Stderr, "aviso (README):", err)
	}
	fmt.Printf("tabelascaffold: estrutura open-source aplicada em %s\n", abs)
	fmt.Printf("  nome=%s org=%s lib=%v stack=%s\n", p.Name, p.Org, p.Lib, p.Stack)
	fmt.Println("  agora: tabelascaffold release . --version v0.1.0")
	return 0
}

// runDoctor takes the same flags as setup — it has to render the same files to
// compare against them — but never writes.
func runDoctor(args []string) int {
	fs := flag.NewFlagSet("doctor", flag.ExitOnError)
	var name, title, org, stack string
	lib := fs.Bool("lib", false, "")
	fs.StringVar(&name, "name", "", "")
	fs.StringVar(&title, "title", "", "")
	fs.StringVar(&org, "org", "", "")
	fs.StringVar(&stack, "stack", "tui", "")

	dir, rest := splitArgs(args, map[string]bool{
		"-name": true, "--name": true,
		"-title": true, "--title": true,
		"-org": true, "--org": true,
		"-stack": true, "--stack": true,
	})
	fs.Parse(rest)
	if fs.NArg() > 0 {
		dir = fs.Arg(0)
	}

	if stack != "tui" && stack != "web" {
		fmt.Fprintf(os.Stderr, "erro: stack desconhecido %q (esperado tui ou web)\n", stack)
		return 1
	}

	abs, err := filepath.Abs(dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		return 1
	}
	if name == "" {
		name = filepath.Base(abs)
	}

	drifts, ign, err := doctor(abs, project{Name: name, Title: title, Org: org, Lib: *lib, Stack: stack})
	if err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		return 1
	}

	if len(drifts) == 0 {
		fmt.Printf("tabelascaffold: %s está alinhado com os templates canônicos\n", abs)
	} else {
		fmt.Printf("tabelascaffold: %d divergência(s) em %s\n", len(drifts), abs)
		for _, d := range drifts {
			fmt.Printf("  %-42s %s\n", d.Path, d.Reason)
		}
	}
	if len(ign) > 0 {
		fmt.Printf("  (%d isento(s) via %s: %s)\n", len(ign), ignoreFile, strings.Join(ign.sorted(), ", "))
	}
	if len(drifts) > 0 {
		return 1
	}
	return 0
}

func runRelease(args []string) int {
	fs := flag.NewFlagSet("release", flag.ExitOnError)
	version := fs.String("version", "", "")
	dir, rest := splitArgs(args, map[string]bool{"-version": true, "--version": true})
	fs.Parse(rest)
	if *version == "" {
		fmt.Fprintln(os.Stderr, "tabelascaffold: --version é obrigatório (ex: --version v0.2.0)")
		return 1
	}
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

// splitArgs separates the one positional dir from the flag args. Go's flag
// package stops parsing at the first non-flag argument, so a trailing
// "--version" (or any flag) after the dir would be silently ignored; this
// pulls the dir out first and preserves flag/value pairs no matter where the
// dir appears.
func splitArgs(args []string, valueFlags map[string]bool) (dir string, rest []string) {
	dir = "."
	seenDir := false
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
	return dir, rest
}
