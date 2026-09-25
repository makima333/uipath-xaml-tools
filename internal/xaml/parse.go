package xaml

import (
	"encoding/xml"
	"fmt"
	"io"
)

func Parse(r io.Reader) (*Node, error) {
	decoder := xml.NewDecoder(r)
	var root *Node
	var stack []*Node
	var childCounts []map[string]int

	for {
		tok, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("parse XAML: %w", err)
		}

		switch token := tok.(type) {
		case xml.StartElement:
			node := &Node{Name: token.Name, Attrs: append([]xml.Attr(nil), token.Attr...)}
			if len(stack) == 0 {
				node.Path = "/" + token.Name.Local + "[1]"
				root = node
			} else {
				parent := stack[len(stack)-1]
				counts := childCounts[len(childCounts)-1]
				counts[token.Name.Local]++
				node.Path = fmt.Sprintf("%s/%s[%d]", parent.Path, token.Name.Local, counts[token.Name.Local])
				parent.Children = append(parent.Children, node)
			}
			stack = append(stack, node)
			childCounts = append(childCounts, map[string]int{})
		case xml.CharData:
			if len(stack) > 0 {
				stack[len(stack)-1].Text += string(token)
			}
		case xml.EndElement:
			if len(stack) == 0 {
				return nil, fmt.Errorf("parse XAML: unexpected closing element %s", token.Name.Local)
			}
			stack = stack[:len(stack)-1]
			childCounts = childCounts[:len(childCounts)-1]
		}
	}

	if root == nil {
		return nil, fmt.Errorf("parse XAML: document is empty")
	}
	return root, nil
}
