package mdast

import (
	"fmt"
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
	indentStr := strings.Repeat("  ", indent)

	// Print node type with additional information
	switch node := n.(type) {
	case *DocumentNode:
		sb.WriteString(fmt.Sprintf("%sDocument\n", indentStr))
	case *HeadingNode:
		sb.WriteString(fmt.Sprintf("%sHeading (level=%d)\n", indentStr, node.Level))
	case *ParagraphNode:
		sb.WriteString(fmt.Sprintf("%sParagraph\n", indentStr))
	case *TextNode:
		sb.WriteString(fmt.Sprintf("%sText: %q\n", indentStr, node.Content))
	default:
		sb.WriteString(fmt.Sprintf("%s%s\n", indentStr, n.Type()))
	}

	// Recursively print children
	for _, c := range n.Children() {
		sb.WriteString(PrintAST(c, indent+1))
	}

	return sb.String()
}

