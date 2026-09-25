package xaml

import (
	"encoding/xml"
	"strings"
)

// Node is a namespace-aware, source-independent representation of an XML node.
// Attribute order is retained to keep diagnostics and tests deterministic.
type Node struct {
	Name     xml.Name
	Attrs    []xml.Attr
	Text     string
	Children []*Node
	Path     string
}

func (n *Node) Local() string { return n.Name.Local }

func (n *Node) Attr(local string) (string, bool) {
	for _, attr := range n.Attrs {
		if attr.Name.Local == local {
			return attr.Value, true
		}
	}
	return "", false
}

func (n *Node) TrimmedText() string { return strings.TrimSpace(n.Text) }

func (n *Node) Walk(fn func(*Node)) {
	fn(n)
	for _, child := range n.Children {
		child.Walk(fn)
	}
}
