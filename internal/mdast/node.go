package mdast

type NodeType string

const (
	NodeDocument  NodeType = "Document"
	NodeHeading   NodeType = "Heading"
	NodeParagraph NodeType = "Paragraph"
	NodeText      NodeType = "Text"
)

type Node interface {
	Type() NodeType
	Children() []Node
	AddChild(Node)
}

type BaseNode struct {
	nodeType NodeType
	children []Node
}

func (n *BaseNode) Type() NodeType   { return n.nodeType }
func (n *BaseNode) Children() []Node { return n.children }
func (n *BaseNode) AddChild(c Node)  { 
	// Pre-allocate capacity to reduce reallocations
	if n.children == nil {
		n.children = make([]Node, 0, 8)
	}
	n.children = append(n.children, c) 
}

type DocumentNode struct {
	BaseNode
}

type HeadingNode struct {
	BaseNode
	Level int
}

type ParagraphNode struct {
	BaseNode
}

type TextNode struct {
	BaseNode
	Content string
}
