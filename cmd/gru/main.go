// gru is the Gru packaging CLI — build, test helpers, and package desktop apps.
//
// Usage:
//
//	gru version
//	gru build hello|calc|leandemo|demo|notepad [windows|linux] [-o PATH] [-webview2]
//	gru package hello|calc|leandemo|demo|notepad [windows|linux]
//	gru icons regen
//	gru fonts convert INPUT.woff2 [-o OUT.ttf]
//	gru check [-skip-vuln] [-skip-lint] [-skip-full-test] [-full-lint]
package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage(os.Stderr)
		os.Exit(2)
	}
	switch os.Args[1] {
	case "version", "-version", "--version":
		printVersion()
	case "build":
		if err := runBuild(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "gru build: %v\n", err)
			os.Exit(1)
		}
	case "new":
		if err := runNew(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "gru new: %v\n", err)
			os.Exit(1)
		}
	case "package":
		if err := runPackage(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "gru package: %v\n", err)
			os.Exit(1)
		}
	case "icons":
		if err := runIcons(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "gru icons: %v\n", err)
			os.Exit(1)
		}
	case "fonts":
		if err := runFonts(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "gru fonts: %v\n", err)
			os.Exit(1)
		}
	case "check":
		if err := runCheck(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "gru check: %v\n", err)
			os.Exit(1)
		}
	case "help", "-h", "--help":
		printUsage(os.Stdout)
	default:
		fmt.Fprintf(os.Stderr, "gru: unknown command %q\n", os.Args[1])
		printUsage(os.Stderr)
		os.Exit(2)
	}
}

func printUsage(w *os.File) {
	fmt.Fprintf(w, `gru — Gru packaging CLI

Usage:
  gru version
  gru new <name> [--webview]           scaffold native or WebView hello starter
  gru build hello    [windows|linux] [-o PATH]
  gru build calc     [windows|linux] [-o PATH]   (quality calculator canary)
  gru build leandemo [windows|linux] [-o PATH]   (thin public scene host)
  gru build webviewhello [windows|linux] [-o PATH] [-webview2]
  gru build demo     [windows|linux] [-o PATH] [-webview2]
  gru build notepad  [windows|linux] [-o PATH]   (private Notepad harness)
  gru package hello|calc|leandemo|webviewhello|demo|notepad [windows|linux]
  gru icons regen
  gru fonts convert INPUT.woff2 [-o OUT.ttf]
  gru check [-skip-vuln] [-skip-lint] [-skip-full-test] [-full-lint]

Apps:
  hello         native public sample (./cmd/hello)
  webviewhello  WebView shell starter (./cmd/webviewhello)
  calc          quality calculator — docs/LEAN_GRU_SPIKE.md
  leandemo      thin host: Counter/Form/ListTile/Toggle/Checkbox (leanpublic)
  demo          curated catalog (-tags grudemo)
  notepad       private product (-tags notepad)

Dev loop:
  go run ./cmd/hello
  go run ./cmd/webviewhello
  go run -tags webview2 ./cmd/webviewhello
  go run ./cmd/calc
  go run ./cmd/leandemo
  go run -tags grudemo .

Start here: docs/CAPABILITY_GUIDE.md

`)
}
