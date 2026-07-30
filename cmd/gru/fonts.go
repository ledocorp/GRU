package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tdewolff/font"
)

// runFonts handles: gru fonts convert
func runFonts(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: gru fonts convert INPUT.woff2 [-o OUT.ttf]")
	}
	switch args[0] {
	case "convert":
		return runFontsConvert(args[1:])
	default:
		return fmt.Errorf("usage: gru fonts convert INPUT.woff2 [-o OUT.ttf]")
	}
}

func runFontsConvert(args []string) error {
	fs := flag.NewFlagSet("fonts convert", flag.ContinueOnError)
	outPath := fs.String("o", "", "output .ttf or .otf path (default: same basename as input)")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: gru fonts convert INPUT.woff2 [-o OUT.ttf]")
	}
	inPath := fs.Arg(0)
	inData, err := os.ReadFile(inPath)
	if err != nil {
		return err
	}
	sfnt, err := font.ToSFNT(inData)
	if err != nil {
		return fmt.Errorf("fonts convert: %w", err)
	}
	ext := font.Extension(sfnt)
	if ext == "" {
		ext = ".ttf"
	}
	dest := *outPath
	if dest == "" {
		base := strings.TrimSuffix(inPath, filepath.Ext(inPath))
		dest = base + ext
	} else if filepath.Ext(dest) == "" {
		dest += ext
	}
	if err := os.WriteFile(dest, sfnt, 0o644); err != nil {
		return err
	}
	fmt.Printf("fonts: wrote %s (%d bytes)\n", dest, len(sfnt))
	fmt.Println("runtime loads TTF/OTF only — commit the .ttf and wire codepoints (see docs/ICONS.md)")
	return nil
}
