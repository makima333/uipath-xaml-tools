package gitdiff_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/makima333/uipath-xaml-tools/internal/gitdiff"
)

func TestGitDiffIntegration(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("Git is not installed")
	}
	dir := t.TempDir()
	// Exercise the real Git shell with spaces and a quote in the executable path.
	binDir := filepath.Join(dir, "uix's tools")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	bin := filepath.Join(binDir, "uix")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	build := exec.Command("go", "build", "-buildvcs=false", "-o", bin, "./cmd/uix")
	build.Dir = filepath.Join("..", "..")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, output)
	}
	repo := filepath.Join(dir, "sample project")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatal(err)
	}
	git := func(args ...string) string {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", repo}, args...)...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, output)
		}
		return string(output)
	}
	write := func(path, data string) {
		t.Helper()
		if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	git("init")
	git("config", "core.autocrlf", "false")
	git("config", "commit.gpgsign", "false")
	workflow := filepath.Join(repo, "Main file.XAML")
	original := `<Sequence xmlns:sap="http://schemas.microsoft.com/netfx/2009/xaml/activities/presentation" sap:VirtualizedContainerService.HintSize="100,100"><WriteLine Text="before" /></Sequence>`
	write(workflow, original)
	git("add", ".")
	git("-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "-m", "Initial")
	attrs := filepath.Join(repo, ".git", "info", "attributes")
	write(attrs, "*.txt -diff\n*.XAML diff=other")
	setup := func(directory string) {
		t.Helper()
		cmd := exec.Command(bin, "setup", "git", directory)
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("setup: %v\n%s", err, output)
		}
	}
	setup(repo)
	first, err := os.ReadFile(attrs)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(first), "*.txt -diff\n*.XAML diff=other\n") {
		t.Fatalf("existing attributes were changed: %s", first)
	}
	subdir := filepath.Join(repo, "nested")
	if err := os.Mkdir(subdir, 0o755); err != nil {
		t.Fatal(err)
	}
	setup(subdir)
	second, err := os.ReadFile(attrs)
	if err != nil || string(first) != string(second) {
		t.Fatalf("setup is not idempotent: %v\n%s", err, second)
	}
	if status := git("status", "--porcelain"); status != "" {
		t.Fatalf("setup changed tracked files: %s", status)
	}
	metadataOnly := strings.Replace(original, "100,100", "200,200", 1)
	write(workflow, metadataOnly)
	if diff := git("diff", "--no-ext-diff"); diff != "" {
		t.Fatalf("metadata-only change should be hidden: %s", diff)
	}
	if diff := git("diff", "--no-ext-diff", "--no-textconv"); !strings.Contains(diff, "200,200") {
		t.Fatalf("raw diff missing metadata: %s", diff)
	}
	write(workflow, strings.Replace(metadataOnly, "before", "after", 1))
	checkDiff := func(diff string) {
		t.Helper()
		if !strings.Contains(diff, "before") || !strings.Contains(diff, "after") || strings.Contains(diff, "HintSize") || strings.Contains(diff, "<Sequence") {
			t.Fatalf("expected compact semantic diff: %s", diff)
		}
	}
	checkDiff(git("diff", "--no-ext-diff"))
	git("add", ".")
	checkDiff(git("diff", "--no-ext-diff", "--cached"))
	if blob := git("show", "--textconv", "HEAD:Main file.XAML"); !strings.Contains(blob, "before") || strings.Contains(blob, "<Sequence") {
		t.Fatalf("historical blob was not converted: %s", blob)
	}
	// textconv must not apply compact's filename-extension restriction.
	temporary := filepath.Join(dir, "blob-without-extension")
	write(temporary, original)
	if output, err := exec.Command(bin, "textconv", temporary).CombinedOutput(); err != nil || !strings.Contains(string(output), "before") {
		t.Fatalf("extensionless blob: %v\n%s", err, output)
	}
}

func TestSetupRejectsNonRepository(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("Git is not installed")
	}
	dir := t.TempDir()
	if err := gitdiff.Setup(dir, filepath.Join(dir, "uix")); err == nil {
		t.Fatal("expected an error outside a Git repository")
	}
	if _, err := os.Stat(filepath.Join(dir, ".git")); !os.IsNotExist(err) {
		t.Fatalf("setup unexpectedly created a repository: %v", err)
	}
}
