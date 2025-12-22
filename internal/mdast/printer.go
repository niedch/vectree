package mdast

import (
	"strconv"
	"strings"
)

// PrintAST prints the AST structure for debugging purposes
func PrintAST(n Node, indent int) string {
	var sb strings.Builder
	printASTHelper(n, indent, &sb)
	return sb.String()
}

func printASTHelper(n Node, indent int, sb *strings.Builder) {
	// Write indentation directly without allocating a string
	for range indent {
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

// NodeToMarkdown converts a node and its children back to markdown format
func NodeToMarkdown(n Node) string {
	var sb strings.Builder
	nodeToMarkdownHelper(n, &sb)
	return sb.String()
}

// nodeToMarkdown is an exported wrapper for the internal helper
func nodeToMarkdown(n Node) string {
	return NodeToMarkdown(n)
}

func nodeToMarkdownHelper(n Node, sb *strings.Builder) {
	switch node := n.(type) {
	case *DocumentNode:
		// Document node just renders its children
		for i, child := range node.Children() {
			nodeToMarkdownHelper(child, sb)
			// Add blank line between top-level blocks (except after last child)
			if i < len(node.Children())-1 {
				sb.WriteByte('\n')
			}
		}

	case *HeadingNode:
		// Render heading with appropriate number of # symbols
		for range node.Level {
			sb.WriteByte('#')
		}
		sb.WriteByte(' ')
		// Render heading content (typically text nodes)
		for _, child := range node.Children() {
			nodeToMarkdownHelper(child, sb)
		}
		sb.WriteByte('\n')

	case *ParagraphNode:
		// Render paragraph content
		for _, child := range node.Children() {
			nodeToMarkdownHelper(child, sb)
		}
		sb.WriteByte('\n')

	case *TextNode:
		// Render text content directly
		sb.WriteString(node.Content)

	default:
		// Fallback for unknown node types
		for _, child := range n.Children() {
			nodeToMarkdownHelper(child, sb)
		}
	}
}
