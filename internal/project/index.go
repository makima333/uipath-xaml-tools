package project

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/makima333/uipath-xaml-tools/internal/workspace"
	"github.com/makima333/uipath-xaml-tools/internal/xaml"
)

type Argument struct {
	Direction string `json:"direction"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Default   string `json:"default,omitempty"`
}

type Workflow struct {
	Path          string     `json:"path"`
	Bytes         int64      `json:"bytes"`
	ActivityCount int        `json:"activityCount"`
	Arguments     []Argument `json:"arguments,omitempty"`
	Invokes       []string   `json:"invokes,omitempty"`
	DynamicInvoke bool       `json:"dynamicInvoke,omitempty"`
	ParseError    string     `json:"parseError,omitempty"`
}

type Index struct {
	Name         string     `json:"name"`
	Main         string     `json:"main,omitempty"`
	ProjectPath  string     `json:"projectPath"`
	Dependencies []string   `json:"dependencies,omitempty"`
	Workflows    []Workflow `json:"workflows"`
}

var argumentType = regexp.MustCompile(`^(In|Out|InOut)Argument\((.+)\)$`)

func Analyze(projectDir string) (Index, error) {
	base, files, err := workspace.XAMLFiles(projectDir)
	if err != nil {
		return Index{}, err
	}
	result := Index{Name: filepath.Base(base), ProjectPath: filepath.ToSlash(base)}
	readProjectJSON(base, &result)
	for _, file := range files {
		result.Workflows = append(result.Workflows, analyzeWorkflow(base, file))
	}
	return result, nil
}

func readProjectJSON(base string, index *Index) {
	data, err := os.ReadFile(filepath.Join(base, "project.json"))
	if err != nil {
		return
	}
	var raw struct {
		Name         string                     `json:"name"`
		Main         string                     `json:"main"`
		Dependencies map[string]json.RawMessage `json:"dependencies"`
	}
	if json.Unmarshal(data, &raw) != nil {
		return
	}
	if raw.Name != "" {
		index.Name = raw.Name
	}
	index.Main = normalizePath(raw.Main)
	for name := range raw.Dependencies {
		index.Dependencies = append(index.Dependencies, name)
	}
	sort.Strings(index.Dependencies)
}

func analyzeWorkflow(base, path string) Workflow {
	info, _ := os.Stat(path)
	workflow := Workflow{Path: workspace.Relative(base, path)}
	if info != nil {
		workflow.Bytes = info.Size()
	}
	file, err := os.Open(path)
	if err != nil {
		workflow.ParseError = err.Error()
		return workflow
	}
	defer file.Close()
	root, err := xaml.Parse(file)
	if err != nil {
		workflow.ParseError = err.Error()
		return workflow
	}

	defaults := argumentDefaults(root)
	visitWorkflow(root, &workflow, defaults)
	return workflow
}

func visitWorkflow(node *xaml.Node, workflow *Workflow, defaults map[string]string) {
	name := node.Local()
	if isIndexNoise(name) {
		return
	}
	if name == "Property" {
		typeName, _ := node.Attr("Type")
		match := argumentType.FindStringSubmatch(typeName)
		if len(match) == 3 {
			argName, _ := node.Attr("Name")
			workflow.Arguments = append(workflow.Arguments, Argument{
				Direction: strings.ToLower(match[1]), Name: argName, Type: match[2], Default: defaults[argName],
			})
		}
	} else if !strings.Contains(name, ".") && name != "Variable" && name != "Members" && name != "Activity" {
		workflow.ActivityCount++
		if name == "InvokeWorkflowFile" {
			target, ok := node.Attr("WorkflowFileName")
			if !ok || target == "" || strings.Contains(target, "[") {
				workflow.DynamicInvoke = true
			} else {
				target = normalizePath(target)
				found := false
				for _, current := range workflow.Invokes {
					if strings.EqualFold(current, target) {
						found = true
						break
					}
				}
				if !found {
					workflow.Invokes = append(workflow.Invokes, target)
				}
			}
		}
	}
	for _, child := range node.Children {
		visitWorkflow(child, workflow, defaults)
	}
}

func isIndexNoise(name string) bool {
	switch name {
	case "WorkflowViewStateService.ViewState",
		"VirtualizedContainerService.HintSize",
		"WorkflowViewState.IdRef",
		"TextExpression.NamespacesForImplementation",
		"TextExpression.ReferencesForImplementation",
		"TextExpression.Namespaces",
		"TextExpression.References":
		return true
	default:
		return false
	}
}

func argumentDefaults(root *xaml.Node) map[string]string {
	result := map[string]string{}
	for _, attr := range root.Attrs {
		if strings.HasPrefix(attr.Name.Space, "clr-namespace") && strings.Contains(attr.Name.Local, ".") {
			parts := strings.Split(attr.Name.Local, ".")
			result[parts[len(parts)-1]] = attr.Value
		}
	}
	return result
}

func normalizePath(value string) string {
	value = strings.ReplaceAll(value, "\\", "/")
	return strings.TrimPrefix(value, "./")
}

func EncodeJSON(index Index) ([]byte, error) {
	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode index: %w", err)
	}
	return append(data, '\n'), nil
}
