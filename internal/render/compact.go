package render

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/makima333/uipath-xaml-tools/internal/ir"
)

type CompactOptions struct {
	Refs bool
}

func Compact(root *ir.Node, options CompactOptions) string {
	var lines []string
	renderNode(root, 0, options, &lines)
	return strings.Join(lines, "\n") + "\n"
}

func renderNode(node *ir.Node, depth int, options CompactOptions, lines *[]string) {
	if node == nil {
		return
	}
	pad := strings.Repeat("  ", depth)
	if node.Annotation != "" {
		for _, line := range strings.Split(node.Annotation, "\n") {
			*lines = append(*lines, pad+"# "+strings.TrimSpace(line))
		}
	}
	if node.Kind == "Property" {
		renderProperty(node, depth, options, lines)
		return
	}

	head, consumed := activityHead(node)
	if node.DisplayName != "" && node.DisplayName != node.Kind {
		head += " " + strconv.Quote(node.DisplayName)
	}
	if fields := renderFieldsExcept(node.Fields, consumed); fields != "" {
		head += " [" + fields + "]"
	}
	if node.Disabled {
		head += " [DISABLED]"
	}
	if options.Refs && node.Ref != "" {
		head += " #" + node.Ref
	}
	*lines = append(*lines, pad+head)

	for _, child := range node.Children {
		if consumed[child.Role] {
			continue
		}
		renderNode(child, depth+1, options, lines)
	}
}

func activityHead(node *ir.Node) (string, map[string]bool) {
	consumed := map[string]bool{}
	switch node.Kind {
	case "Assign":
		to := propertyValue(node, "To")
		value := propertyValue(node, "Value")
		if to != "" || value != "" {
			consumed["To"], consumed["Value"] = true, true
			return fmt.Sprintf("Assign %s = %s", to, value), consumed
		}
	case "If":
		condition := fieldOrProperty(node, "Condition")
		if condition != "" {
			consumed["Condition"] = true
			return "If " + condition, consumed
		}
	case "InvokeWorkflowFile":
		if target := fieldOrProperty(node, "WorkflowFileName"); target != "" {
			consumed["WorkflowFileName"] = true
			return "Invoke " + target, consumed
		}
	case "LogMessage":
		level := fieldOrProperty(node, "Level")
		message := fieldOrProperty(node, "Message")
		if level != "" || message != "" {
			consumed["Level"], consumed["Message"] = true, true
			return strings.TrimSpace("Log " + level + " " + message), consumed
		}
	case "GetAsset":
		asset := fieldOrProperty(node, "AssetName")
		result := fieldOrProperty(node, "Result")
		if asset != "" || result != "" {
			consumed["AssetName"], consumed["Result"] = true, true
			return strings.TrimSpace("GetAsset " + asset + " -> " + result), consumed
		}
	}

	head := node.Kind
	if node.Text != "" {
		head += " " + inline(node.Text)
	}
	return head, consumed
}

func renderProperty(node *ir.Node, depth int, options CompactOptions, lines *[]string) {
	pad := strings.Repeat("  ", depth)
	if value := simpleValue(node); value != "" {
		*lines = append(*lines, pad+node.Role+": "+value)
		return
	}
	*lines = append(*lines, pad+node.Role+":")
	for _, child := range node.Children {
		renderNode(child, depth+1, options, lines)
	}
}

func fieldOrProperty(node *ir.Node, name string) string {
	if value, ok := node.Field(name); ok {
		return value
	}
	return propertyValue(node, name)
}

func propertyValue(node *ir.Node, role string) string {
	if child := node.ChildRole(role); child != nil {
		return simpleValue(child)
	}
	return ""
}

func simpleValue(node *ir.Node) string {
	if node.Text != "" {
		return inline(node.Text)
	}
	if len(node.Children) == 1 {
		child := node.Children[0]
		if child.Text != "" && len(child.Children) == 0 {
			return inline(child.Text)
		}
		if selector, ok := child.Field("Selector"); ok {
			return selector
		}
	}
	if len(node.Fields) == 1 {
		return node.Fields[0].Value
	}
	return ""
}

func inline(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

func renderFields(fields []ir.Field) string {
	parts := make([]string, 0, len(fields))
	for _, field := range fields {
		parts = append(parts, field.Name+"="+strconv.Quote(field.Value))
	}
	return strings.Join(parts, " ")
}

func renderFieldsExcept(fields []ir.Field, excluded map[string]bool) string {
	kept := make([]ir.Field, 0, len(fields))
	for _, field := range fields {
		if !excluded[field.Name] {
			kept = append(kept, field)
		}
	}
	return renderFields(kept)
}
