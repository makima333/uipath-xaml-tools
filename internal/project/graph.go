package project

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

func GraphText(index Index) string {
	var lines []string
	if index.Main != "" {
		lines = append(lines, "main: "+index.Main)
	}
	for _, workflow := range index.Workflows {
		if len(workflow.Invokes) == 0 && !workflow.DynamicInvoke {
			lines = append(lines, workflow.Path)
			continue
		}
		for _, target := range workflow.Invokes {
			lines = append(lines, workflow.Path+" -> "+target)
		}
		if workflow.DynamicInvoke {
			lines = append(lines, workflow.Path+" -> <dynamic>")
		}
	}
	return strings.Join(lines, "\n") + "\n"
}

func GraphMermaid(index Index) string {
	paths := map[string]bool{}
	for _, workflow := range index.Workflows {
		paths[workflow.Path] = true
		for _, target := range workflow.Invokes {
			paths[target] = true
		}
	}
	ordered := make([]string, 0, len(paths))
	for path := range paths {
		ordered = append(ordered, path)
	}
	sort.Strings(ordered)
	ids := map[string]string{}
	var lines = []string{"flowchart TD"}
	for i, path := range ordered {
		id := fmt.Sprintf("W%d", i+1)
		ids[path] = id
		lines = append(lines, fmt.Sprintf("  %s[%s]", id, strconv.Quote(path)))
	}
	for _, workflow := range index.Workflows {
		for _, target := range workflow.Invokes {
			lines = append(lines, fmt.Sprintf("  %s --> %s", ids[workflow.Path], ids[target]))
		}
		if workflow.DynamicInvoke {
			lines = append(lines, fmt.Sprintf("  %s -.-> DYNAMIC((dynamic))", ids[workflow.Path]))
		}
	}
	return strings.Join(lines, "\n") + "\n"
}
