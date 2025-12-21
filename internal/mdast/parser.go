package mdast

import (
	"strconv"
	"strings"
)

func ParseMarkdown(markdown string) *DocumentNode {
	tokens := Lex(markdown);
	return Parse(tokens)
}

func Parse(tokens []Token) *DocumentNode {
	doc := &DocumentNode{
		BaseNode: BaseNode{nodeType: NodeDocument},
	}

	i := 0
	for i < len(tokens) {
		tok := tokens[i]

		switch tok.Type {

		case TokenHeading:
			h := &HeadingNode{
				BaseNode: BaseNode{nodeType: NodeHeading},
				Level:    tok.Level,
			}
			h.AddChild(&TextNode{
				BaseNode: BaseNode{nodeType: NodeText},
				Content:  tok.Value,
			})
			doc.AddChild(h)
			i++

		case TokenText:
			p := &ParagraphNode{
				BaseNode: BaseNode{nodeType: NodeParagraph},
			}

			for i < len(tokens) && tokens[i].Type == TokenText {
				p.AddChild(&TextNode{
					BaseNode: BaseNode{nodeType: NodeText},
					Content:  tokens[i].Value,
				})
				i++
			}

			doc.AddChild(p)

		default:
			i++
		}
	}

	return doc
}

func PrintAST(n Node, indent int) string {
	var sb strings.Builder
	printASTHelper(n, indent, &sb)
	return sb.String()
}

func printASTHelper(n Node, indent int, sb *strings.Builder) {
	// Write indentation directly without allocating a string
	for i := 0; i < indent; i++ {
		sb.WriteString("  ")
	}

	// Print node type with additional information
	// Avoid fmt.Sprintf by building strings directly
	switch node := n.(type) {
	case *DocumentNode:
		sb.WriteString("Document\n")
	case *HeadingNode:
		sb.WriteString("Heading (level=")
		sb.WriteString(strconv.Itoa(node.Level))
		sb.WriteString(")\n")
	case *ParagraphNode:
		sb.WriteString("Paragraph\n")
	case *TextNode:
		sb.WriteString("Text: \"")
		sb.WriteString(node.Content)
		sb.WriteString("\"\n")
	default:
		sb.WriteString(string(n.Type()))
		sb.WriteByte('\n')
	}

	// Recursively print children
	for _, c := range n.Children() {
		printASTHelper(c, indent+1, sb)
	}
}

