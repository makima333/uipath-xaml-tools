package ir

type Field struct {
	Name  string
	Value string
}

// Node is the compact semantic form shared by all renderers.
type Node struct {
	Kind        string
	Namespace   string
	Role        string
	DisplayName string
	Ref         string
	Annotation  string
	Text        string
	Fields      []Field
	Children    []*Node
	Disabled    bool
}

func (n *Node) Field(name string) (string, bool) {
	for _, field := range n.Fields {
		if field.Name == name {
			return field.Value, true
		}
	}
	return "", false
}

func (n *Node) ChildRole(role string) *Node {
	for _, child := range n.Children {
		if child.Role == role {
			return child
		}
	}
	return nil
}
