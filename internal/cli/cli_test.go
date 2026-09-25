package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/makima333/uipath-xaml-tools/internal/cli"
)

func TestCompactCommand(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "Main.xaml")
	if err := os.WriteFile(path, []byte(`<Sequence DisplayName="Main"><WriteLine Text="hello" /></Sequence>`), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if err := cli.Run([]string{"compact", path}, &stdout, &stderr, "test"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), `Sequence "Main"`) || !strings.Contains(stdout.String(), `WriteLine`) {
		t.Fatalf("unexpected output: %s", stdout.String())
	}
}

func TestStripRefusesInputOverwrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "Main.xaml")
	if err := os.WriteFile(path, []byte(`<Sequence />`), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	err := cli.Run([]string{"strip", "-o", dir, path}, &stdout, &stderr, "test")
	if err == nil || !strings.Contains(err.Error(), "refusing to overwrite") {
		t.Fatalf("expected overwrite refusal, got %v", err)
	}
}

func TestSubcommandHelpIsNotAnError(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if err := cli.Run([]string{"strip", "-h"}, &stdout, &stderr, "test"); err != nil {
		t.Fatalf("help returned an error: %v", err)
	}
	if !strings.Contains(stderr.String(), "Usage of strip:") {
		t.Fatalf("help output missing usage: %s", stderr.String())
	}
}
