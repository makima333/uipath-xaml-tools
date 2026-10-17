// Package gitdiff configures repository-local Git text conversion for XAML.
package gitdiff

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const attributeRule = "*.[xX][aA][mM][lL] diff=uix"

// Setup registers the installed executable and an untracked local attribute rule.
// Git invokes textconv through its shell, even on Windows.
func Setup(directory, executable string) error {
	git := func(args ...string) (string, error) {
		cmd := exec.Command("git", append([]string{"-C", directory}, args...)...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(string(output)))
		}
		return strings.TrimSpace(string(output)), nil
	}
	inside, err := git("rev-parse", "--is-inside-work-tree")
	if err != nil {
		return err
	}
	if inside != "true" {
		return fmt.Errorf("setup requires a Git working tree")
	}
	attributes, err := git("rev-parse", "--path-format=absolute", "--git-path", "info/attributes")
	if err != nil {
		return err
	}
	data, err := os.ReadFile(attributes)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	abs, err := filepath.Abs(executable)
	if err != nil {
		return err
	}
	// Single-quote shell metacharacters and escape embedded single quotes.
	command := "'" + strings.ReplaceAll(filepath.ToSlash(abs), "'", "'\"'\"'") + "' textconv"
	if _, err := git("config", "--local", "--replace-all", "diff.uix.textconv", command); err != nil {
		return err
	}
	// Avoid retaining stale IR when the executable is upgraded in place.
	if _, err := git("config", "--local", "--replace-all", "diff.uix.cachetextconv", "false"); err != nil {
		return err
	}
	// Append after existing rules to override other XAML diff drivers locally.
	// Repeated setup without intervening rules leaves the file unchanged.
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if strings.TrimSpace(lines[len(lines)-1]) != attributeRule {
		if len(data) > 0 && data[len(data)-1] != '\n' {
			data = append(data, '\n')
		}
		data = append(data, []byte(attributeRule+"\n")...)
		if err := os.MkdirAll(filepath.Dir(attributes), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(attributes, data, 0o644); err != nil {
			return err
		}
	}
	return nil
}
