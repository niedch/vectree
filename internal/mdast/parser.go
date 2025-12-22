package mdast

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

