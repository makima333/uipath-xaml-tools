package cli

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/makima333/uipath-xaml-tools/internal/gitdiff"
	"github.com/makima333/uipath-xaml-tools/internal/normalize"
	"github.com/makima333/uipath-xaml-tools/internal/project"
	"github.com/makima333/uipath-xaml-tools/internal/render"
	stripper "github.com/makima333/uipath-xaml-tools/internal/strip"
	"github.com/makima333/uipath-xaml-tools/internal/workspace"
	"github.com/makima333/uipath-xaml-tools/internal/xaml"
)

const usage = `uix: UiPath XAML tools for humans and LLMs

Usage:
  uix compact [--refs] [-o DIR] PATH...
  uix strip [--conservative] [-o DIR] PATH...
  uix index [-o FILE] PROJECT_DIR
  uix read [--refs=false] PROJECT_DIR WORKFLOW
  uix graph [--format text|mermaid] PROJECT_DIR
  uix setup git [REPOSITORY_DIR]
  uix textconv FILE
  uix version
`

func Run(args []string, stdout, stderr io.Writer, version string) error {
	if len(args) == 0 {
		fmt.Fprint(stderr, usage)
		return nil
	}
	switch args[0] {
	case "version", "--version":
		_, err := fmt.Fprintf(stdout, "uix %s\n", version)
		return err
	case "help", "--help", "-h":
		_, err := fmt.Fprint(stdout, usage)
		return err
	case "compact":
		return runCompact(args[1:], stdout, stderr)
	case "strip":
		return runStrip(args[1:], stdout, stderr)
	case "index":
		return runIndex(args[1:], stdout, stderr)
	case "read":
		return runRead(args[1:], stdout)
	case "graph":
		return runGraph(args[1:], stdout)
	case "setup":
		return runSetup(args[1:], stdout, stderr)
	case "textconv":
		return runTextconv(args[1:], stdout, stderr)
	default:
		return fmt.Errorf("unknown command %q\n\n%s", args[0], usage)
	}
}

func runSetup(args []string, stdout, stderr io.Writer) error {
	isGit := len(args) > 0 && args[0] == "git"
	if isGit {
		args = args[1:]
	}
	flags := flag.NewFlagSet("setup", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.Usage = func() { fmt.Fprintln(stderr, "Usage: uix setup git [REPOSITORY_DIR]") }
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if !isGit || flags.NArg() > 1 {
		return fmt.Errorf("setup requires: git [REPOSITORY_DIR]")
	}
	directory := "."
	if flags.NArg() == 1 {
		directory = flags.Arg(0)
	}
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	if err := gitdiff.Setup(directory, executable); err != nil {
		return err
	}
	_, err = fmt.Fprintln(stdout, "Configured repository-local XAML diffs with uix. Run git diff or git log -p.\nKeep this uix executable at its current location, or run setup again after moving it.")
	return err
}

func runTextconv(args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("textconv", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.Usage = func() {
		fmt.Fprintln(stderr, "Usage: uix textconv FILE\nConvert one file to compact IR for Git, including temporary files without an extension.")
	}
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() != 1 {
		return fmt.Errorf("textconv requires one file")
	}
	result, err := compactFile(flags.Arg(0), false)
	if err != nil {
		return err
	}
	_, err = io.WriteString(stdout, result)
	return err
}

func runCompact(args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("compact", flag.ContinueOnError)
	flags.SetOutput(stderr)
	outDir := flags.String("o", "", "output directory")
	refs := flags.Bool("refs", false, "include source references")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	return convertPaths(flags.Args(), *outDir, ".uir", stdout, func(path string) (string, error) {
		return compactFile(path, *refs)
	})
}

func runStrip(args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("strip", flag.ContinueOnError)
	flags.SetOutput(stderr)
	outDir := flags.String("o", "", "output directory")
	conservative := flags.Bool("conservative", false, "remove designer-generated data only")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	return convertPaths(flags.Args(), *outDir, ".xaml", stdout, func(path string) (string, error) {
		data, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		return stripper.XAML(string(data), stripper.Options{Conservative: *conservative})
	})
}

func runIndex(args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("index", flag.ContinueOnError)
	flags.SetOutput(stderr)
	output := flags.String("o", "", "output JSON file")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() != 1 {
		return fmt.Errorf("index requires one project directory")
	}
	index, err := project.Analyze(flags.Arg(0))
	if err != nil {
		return err
	}
	data, err := project.EncodeJSON(index)
	if err != nil {
		return err
	}
	return writeOutput(*output, data, stdout)
}

func runRead(args []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("read", flag.ContinueOnError)
	refs := flags.Bool("refs", true, "include source references")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() != 2 {
		return fmt.Errorf("read requires PROJECT_DIR and WORKFLOW")
	}
	path, err := workspace.FindWorkflow(flags.Arg(0), flags.Arg(1))
	if err != nil {
		return err
	}
	result, err := compactFile(path, *refs)
	if err != nil {
		return err
	}
	_, err = io.WriteString(stdout, result)
	return err
}

func runGraph(args []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("graph", flag.ContinueOnError)
	format := flags.String("format", "text", "text or mermaid")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() != 1 {
		return fmt.Errorf("graph requires one project directory")
	}
	index, err := project.Analyze(flags.Arg(0))
	if err != nil {
		return err
	}
	var result string
	switch *format {
	case "text":
		result = project.GraphText(index)
	case "mermaid":
		result = project.GraphMermaid(index)
	default:
		return fmt.Errorf("unknown graph format %q", *format)
	}
	_, err = io.WriteString(stdout, result)
	return err
}

func compactFile(path string, refs bool) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	data = bytes.TrimPrefix(data, []byte{0xef, 0xbb, 0xbf})
	root, err := xaml.Parse(bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("%s: %w", path, err)
	}
	return render.Compact(normalize.Document(root), render.CompactOptions{Refs: refs}), nil
}

func convertPaths(paths []string, outDir, suffix string, stdout io.Writer, convert func(string) (string, error)) error {
	if len(paths) == 0 {
		return fmt.Errorf("at least one XAML file or directory is required")
	}
	type input struct{ base, path string }
	var inputs []input
	for _, path := range paths {
		base, files, err := workspace.XAMLFiles(path)
		if err != nil {
			return err
		}
		for _, file := range files {
			inputs = append(inputs, input{base: base, path: file})
		}
	}
	if len(inputs) == 0 {
		return fmt.Errorf("no XAML files found")
	}

	written := map[string]string{}
	for _, item := range inputs {
		result, err := convert(item.path)
		if err != nil {
			return err
		}
		if outDir == "" {
			if len(inputs) > 1 {
				fmt.Fprintf(stdout, "===== FILE: %s =====\n", workspace.Relative(item.base, item.path))
			}
			if _, err := io.WriteString(stdout, result); err != nil {
				return err
			}
			if !strings.HasSuffix(result, "\n") {
				fmt.Fprintln(stdout)
			}
			continue
		}

		rel := workspace.Relative(item.base, item.path)
		destination := filepath.Join(outDir, filepath.FromSlash(rel))
		destination = strings.TrimSuffix(destination, filepath.Ext(destination)) + suffix
		destAbs, _ := filepath.Abs(destination)
		inputAbs, _ := filepath.Abs(item.path)
		if strings.EqualFold(destAbs, inputAbs) {
			return fmt.Errorf("refusing to overwrite input: %s", item.path)
		}
		key := strings.ToLower(destAbs)
		if previous, ok := written[key]; ok {
			return fmt.Errorf("output collision: %s and %s", previous, item.path)
		}
		written[key] = item.path
		if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(destination, []byte(result), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func writeOutput(path string, data []byte, stdout io.Writer) error {
	if path == "" {
		_, err := stdout.Write(data)
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil && filepath.Dir(path) != "." {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
