package workspace

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func XAMLFiles(path string) (string, []string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", nil, err
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", nil, err
	}
	if !info.IsDir() {
		if !strings.EqualFold(filepath.Ext(abs), ".xaml") {
			return "", nil, fmt.Errorf("not a XAML file: %s", path)
		}
		return filepath.Dir(abs), []string{abs}, nil
	}

	var files []string
	err = filepath.WalkDir(abs, func(current string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() && current != abs && strings.HasPrefix(entry.Name(), ".") {
			return filepath.SkipDir
		}
		if !entry.IsDir() && strings.EqualFold(filepath.Ext(entry.Name()), ".xaml") {
			files = append(files, current)
		}
		return nil
	})
	sort.Strings(files)
	return abs, files, err
}

func Relative(base, path string) string {
	rel, err := filepath.Rel(base, path)
	if err != nil {
		return filepath.ToSlash(path)
	}
	return filepath.ToSlash(rel)
}

func FindWorkflow(projectDir, query string) (string, error) {
	base, files, err := XAMLFiles(projectDir)
	if err != nil {
		return "", err
	}
	query = strings.TrimPrefix(filepath.ToSlash(query), "./")
	var matches []string
	for _, file := range files {
		rel := Relative(base, file)
		if strings.EqualFold(rel, query) || strings.EqualFold(filepath.Base(file), query) {
			matches = append(matches, file)
		}
	}
	if len(matches) == 0 && filepath.Ext(query) == "" {
		return FindWorkflow(projectDir, query+".xaml")
	}
	if len(matches) == 0 {
		return "", fmt.Errorf("workflow not found: %s", query)
	}
	if len(matches) > 1 {
		return "", fmt.Errorf("workflow name is ambiguous: %s", query)
	}
	return matches[0], nil
}
