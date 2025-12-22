package mdast

import (
	"testing"
)

func TestNodeToMarkdown_SimpleHeading(t *testing.T) {
	markdown := "# Hello World"
	doc := ParseMarkdown(markdown)
	
	result := doc.ToMarkdown()
	expected := "# Hello World\n"
	
	if result != expected {
		t.Errorf("Markdown conversion failed.\nExpected:\n%q\nActual:\n%q", expected, result)
	}
}

func TestNodeToMarkdown_SimpleParagraph(t *testing.T) {
	markdown := "This is a paragraph."
	doc := ParseMarkdown(markdown)
	
	result := doc.ToMarkdown()
	expected := "This is a paragraph.\n"
	
	if result != expected {
		t.Errorf("Markdown conversion failed.\nExpected:\n%q\nActual:\n%q", expected, result)
	}
}

func TestNodeToMarkdown_MultiLineParagraph(t *testing.T) {
	markdown := `Line one.
Line two.
Line three.`
	doc := ParseMarkdown(markdown)
	
	result := doc.ToMarkdown()
	expected := "Line one.Line two.Line three.\n"
	
	if result != expected {
		t.Errorf("Markdown conversion failed.\nExpected:\n%q\nActual:\n%q", expected, result)
	}
}

func TestNodeToMarkdown_HeadingAndParagraph(t *testing.T) {
	markdown := `# Introduction

This is the introduction paragraph.`
	doc := ParseMarkdown(markdown)
	
	result := doc.ToMarkdown()
	expected := "# Introduction\n\nThis is the introduction paragraph.\n"
	
	if result != expected {
		t.Errorf("Markdown conversion failed.\nExpected:\n%q\nActual:\n%q", expected, result)
	}
}

func TestNodeToMarkdown_ComplexDocument(t *testing.T) {
	markdown := `# Main Title

First paragraph line one.
First paragraph line two.

## Section One

Section one content.

## Section Two

Section two line one.
Section two line two.`
	
	doc := ParseMarkdown(markdown)
	result := doc.ToMarkdown()
	
	expected := `# Main Title

First paragraph line one.First paragraph line two.

## Section One

Section one content.

## Section Two

Section two line one.Section two line two.
`
	
	if result != expected {
		t.Errorf("Markdown conversion failed.\nExpected:\n%q\nActual:\n%q", expected, result)
	}
}

func TestNodeToMarkdown_AllHeadingLevels(t *testing.T) {
	markdown := `# Level 1
## Level 2
### Level 3
#### Level 4
##### Level 5
###### Level 6`
	
	doc := ParseMarkdown(markdown)
	result := doc.ToMarkdown()
	
	expected := `# Level 1

## Level 2

### Level 3

#### Level 4

##### Level 5

###### Level 6
`
	
	if result != expected {
		t.Errorf("Markdown conversion failed.\nExpected:\n%q\nActual:\n%q", expected, result)
	}
}

func TestNodeToMarkdown_IndividualHeadingNode(t *testing.T) {
	heading := &HeadingNode{
		BaseNode: BaseNode{nodeType: NodeHeading},
		Level:    2,
	}
	heading.AddChild(&TextNode{
		BaseNode: BaseNode{nodeType: NodeText},
		Content:  "Test Heading",
	})
	
	result := heading.ToMarkdown()
	expected := "## Test Heading\n"
	
	if result != expected {
		t.Errorf("Markdown conversion failed.\nExpected:\n%q\nActual:\n%q", expected, result)
	}
}

func TestNodeToMarkdown_IndividualParagraphNode(t *testing.T) {
	para := &ParagraphNode{
		BaseNode: BaseNode{nodeType: NodeParagraph},
	}
	para.AddChild(&TextNode{
		BaseNode: BaseNode{nodeType: NodeText},
		Content:  "Test paragraph.",
	})
	
	result := para.ToMarkdown()
	expected := "Test paragraph.\n"
	
	if result != expected {
		t.Errorf("Markdown conversion failed.\nExpected:\n%q\nActual:\n%q", expected, result)
	}
}

func TestNodeToMarkdown_IndividualTextNode(t *testing.T) {
	text := &TextNode{
		BaseNode: BaseNode{nodeType: NodeText},
		Content:  "Just some text",
	}
	
	result := text.ToMarkdown()
	expected := "Just some text"
	
	if result != expected {
		t.Errorf("Markdown conversion failed.\nExpected:\n%q\nActual:\n%q", expected, result)
	}
}

func TestNodeToMarkdown_EmptyDocument(t *testing.T) {
	doc := &DocumentNode{
		BaseNode: BaseNode{nodeType: NodeDocument},
	}
	
	result := doc.ToMarkdown()
	expected := ""
	
	if result != expected {
		t.Errorf("Markdown conversion failed.\nExpected:\n%q\nActual:\n%q", expected, result)
	}
}
