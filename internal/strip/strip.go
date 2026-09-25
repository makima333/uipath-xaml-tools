package strip

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/makima333/uipath-xaml-tools/internal/xaml"
)

type Options struct {
	Conservative bool
}

var conservativeBlocks = []string{
	"WorkflowViewStateService\\.ViewState",
	"VirtualizedContainerService\\.HintSize",
	"WorkflowViewState\\.IdRef",
}

var aggressiveBlocks = []string{
	"TextExpression\\.NamespacesForImplementation",
	"TextExpression\\.ReferencesForImplementation",
	"TextExpression\\.Namespaces",
	"TextExpression\\.References",
}

var (
	hintSizeAttr = regexp.MustCompile(`\s+(?:[A-Za-z_][\w.-]*:)?VirtualizedContainerService\.HintSize\s*=\s*"[^"]*"`)
	idRefAttr    = regexp.MustCompile(`\s+(?:[A-Za-z_][\w.-]*:)?WorkflowViewState\.IdRef\s*=\s*"[^"]*"`)
	screenshot   = regexp.MustCompile(`\s+(?:InformativeScreenshot|ScreenshotIdentifier)\s*=\s*"[^"]*"`)
	nullAttr     = regexp.MustCompile(`\s+[A-Za-z_][\w.:-]*\s*=\s*"\{x:Null\}"`)
	guidAttr     = regexp.MustCompile(`(?i)\s+(?:Id|Reference)\s*=\s*"[0-9a-f]{8}(?:-[0-9a-f]{4}){3}-[0-9a-f]{12}"`)
	blankLines   = regexp.MustCompile(`\r?\n[ \t]*\r?\n(?:[ \t]*\r?\n)+`)
)

func XAML(input string, options Options) (string, error) {
	result := input
	for _, tag := range conservativeBlocks {
		result = removeBlock(result, tag)
	}
	result = hintSizeAttr.ReplaceAllString(result, "")
	result = idRefAttr.ReplaceAllString(result, "")

	if !options.Conservative {
		for _, tag := range aggressiveBlocks {
			result = removeBlock(result, tag)
		}
		result = screenshot.ReplaceAllString(result, "")
		result = nullAttr.ReplaceAllString(result, "")
		result = guidAttr.ReplaceAllString(result, "")
	}
	result = blankLines.ReplaceAllString(result, "\n")

	if _, err := xaml.Parse(strings.NewReader(result)); err != nil {
		return "", fmt.Errorf("strip produced invalid XML: %w", err)
	}
	return result, nil
}

func removeBlock(input, localPattern string) string {
	qualified := `(?:[A-Za-z_][\w.-]*:)?` + localPattern
	paired := regexp.MustCompile(`(?s)[ \t]*<` + qualified + `(?:\s[^>]*)?>.*?</` + qualified + `\s*>[ \t]*(?:\r?\n)?`)
	selfClosing := regexp.MustCompile(`[ \t]*<` + qualified + `(?:\s[^>]*)?/\s*>[ \t]*(?:\r?\n)?`)
	return selfClosing.ReplaceAllString(paired.ReplaceAllString(input, ""), "")
}
