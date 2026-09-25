package normalize

import (
	"regexp"
	"strings"

	"github.com/makima333/uipath-xaml-tools/internal/ir"
	"github.com/makima333/uipath-xaml-tools/internal/xaml"
)

const (
	sap2010Namespace = "http://schemas.microsoft.com/netfx/2010/xaml/activities/presentation"
	annotationLocal  = "Annotation.AnnotationText"
)

var guidPattern = regexp.MustCompile(`(?i)^[0-9a-f]{8}(?:-[0-9a-f]{4}){3}-[0-9a-f]{12}$`)

var noiseElements = map[string]bool{
	"WorkflowViewStateService.ViewState":         true,
	"VirtualizedContainerService.HintSize":       true,
	"WorkflowViewState.IdRef":                    true,
	"TextExpression.NamespacesForImplementation": true,
	"TextExpression.ReferencesForImplementation": true,
	"TextExpression.Namespaces":                  true,
	"TextExpression.References":                  true,
}

var defaultFields = map[string]string{
	"ClickType":    "CLICK_SINGLE",
	"KeyModifiers": "None",
	"MouseButton":  "BTN_LEFT",
	"Position":     "Center",
	"Ignorable":    "sap sap2010",
}

func Document(root *xaml.Node) *ir.Node { return normalizeNode(root, "") }

func normalizeNode(node *xaml.Node, role string) *ir.Node {
	if noiseElements[node.Local()] || isEmptyScaffold(node) {
		return nil
	}

	out := &ir.Node{
		Kind:      node.Local(),
		Namespace: node.Name.Space,
		Role:      role,
		Ref:       node.Path,
		Text:      node.TrimmedText(),
		Disabled:  node.Local() == "CommentOut",
	}
	if strings.Contains(node.Local(), ".") {
		parts := strings.Split(node.Local(), ".")
		out.Role = parts[len(parts)-1]
		out.Kind = "Property"
	}

	for _, attr := range node.Attrs {
		name := attr.Name.Local
		switch {
		case name == "xmlns" || attr.Name.Space == "xmlns":
			continue
		case name == "DisplayName":
			if attr.Value != "{x:Null}" {
				out.DisplayName = attr.Value
			}
		case attr.Name.Space == sap2010Namespace && name == annotationLocal:
			out.Annotation = strings.TrimSpace(attr.Value)
		case name == "WorkflowViewState.IdRef":
			out.Ref = attr.Value
		case isNoiseAttribute(name, attr.Value):
			continue
		default:
			if strings.HasSuffix(name, "TypeArguments") {
				name = "T"
			}
			out.Fields = append(out.Fields, ir.Field{Name: name, Value: attr.Value})
		}
	}

	for _, child := range node.Children {
		normalized := normalizeNode(child, "")
		if normalized != nil {
			out.Children = append(out.Children, normalized)
		}
	}
	return out
}

func isNoiseAttribute(name, value string) bool {
	if name == "VirtualizedContainerService.HintSize" || name == "InformativeScreenshot" || name == "ScreenshotIdentifier" {
		return true
	}
	if value == "{x:Null}" || defaultFields[name] == value {
		return true
	}
	return (name == "Id" || name == "Reference") && guidPattern.MatchString(value)
}

func isEmptyScaffold(node *xaml.Node) bool {
	name := node.Local()
	if noiseElements[name] {
		return true
	}
	if node.TrimmedText() != "" {
		return false
	}
	if !(strings.HasSuffix(name, "Argument") || name == "CursorPosition" || strings.Contains(name, ".")) {
		return false
	}
	for _, attr := range node.Attrs {
		if !strings.HasSuffix(attr.Name.Local, "TypeArguments") && !isNoiseAttribute(attr.Name.Local, attr.Value) {
			return false
		}
	}
	for _, child := range node.Children {
		if !isEmptyScaffold(child) {
			return false
		}
	}
	return true
}
