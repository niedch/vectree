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
	ToMarkdown() string
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
func (n *BaseNode) ToMarkdown() string { return "" }

type DocumentNode struct {
	BaseNode
}

func (n *DocumentNode) ToMarkdown() string {
	return nodeToMarkdown(n)
}

type HeadingNode struct {
	BaseNode
	Level int
}

func (n *HeadingNode) ToMarkdown() string {
	return nodeToMarkdown(n)
}

type ParagraphNode struct {
	BaseNode
}

func (n *ParagraphNode) ToMarkdown() string {
	return nodeToMarkdown(n)
}

type TextNode struct {
	BaseNode
	Content string
}

func (n *TextNode) ToMarkdown() string {
	return nodeToMarkdown(n)
}
