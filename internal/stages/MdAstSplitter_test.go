package stages

import (
	"context"
	"testing"
	"time"
)

func TestMdAstSplitter_Run(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:  "Single heading",
			input: "# Hello World",
			expected: []string{
				"# Hello World\n",
			},
		},
		{
			name: "Heading with paragraph",
			input: `# Title

This is a paragraph.`,
			expected: []string{
				"# Title\nThis is a paragraph.\n",
			},
		},
		{
			name: "Multiple top-level headings",
			input: `# Title One

Paragraph one.

# Title Two

Paragraph two.`,
			expected: []string{
				"# Title One\nParagraph one.\n",
				"# Title Two\nParagraph two.\n",
			},
		},
		{
			name: "Heading with subheadings - overlapping outputs",
			input: `# Main Title

Introduction paragraph.

## Section One

Section one content.

### Subsection 1.1

Subsection content.

## Section Two

Section two content.`,
			expected: []string{
				"# Main Title\nIntroduction paragraph.\n## Section One\nSection one content.\n### Subsection 1.1\nSubsection content.\n## Section Two\nSection two content.\n",
				"## Section One\nSection one content.\n### Subsection 1.1\nSubsection content.\n",
				"### Subsection 1.1\nSubsection content.\n",
				"## Section Two\nSection two content.\n",
			},
		},
		{
			name: "Multiple level 1 headings with subheadings",
			input: `# First Title

First intro.

## First Section

First section content.

# Second Title

Second intro.

## Second Section

Second section content.`,
			expected: []string{
				"# First Title\nFirst intro.\n## First Section\nFirst section content.\n",
				"## First Section\nFirst section content.\n",
				"# Second Title\nSecond intro.\n## Second Section\nSecond section content.\n",
				"## Second Section\nSecond section content.\n",
			},
		},
		{
			name: "Complex nested structure",
			input: `# Main Title

First paragraph.

## Section One

Content for section one.

### Subsection 1.1

Details here.

### Subsection 1.2

More details.

## Section Two

Content for section two.`,
			expected: []string{
				"# Main Title\nFirst paragraph.\n## Section One\nContent for section one.\n### Subsection 1.1\nDetails here.\n### Subsection 1.2\nMore details.\n## Section Two\nContent for section two.\n",
				"## Section One\nContent for section one.\n### Subsection 1.1\nDetails here.\n### Subsection 1.2\nMore details.\n",
				"### Subsection 1.1\nDetails here.\n",
				"### Subsection 1.2\nMore details.\n",
				"## Section Two\nContent for section two.\n",
			},
		},
		{
			name:     "No headings",
			input:    "Just a paragraph with no headings.",
			expected: []string{},
		},
		{
			name: "Multiple level 2 headings",
			input: `## Section One

Content one.

## Section Two

Content two.

## Section Three

Content three.`,
			expected: []string{
				"## Section One\nContent one.\n",
				"## Section Two\nContent two.\n",
				"## Section Three\nContent three.\n",
			},
		},
		{
			name: "Deep nesting",
			input: `# Level 1

L1 content.

## Level 2

L2 content.

### Level 3

L3 content.

#### Level 4

L4 content.`,
			expected: []string{
				"# Level 1\nL1 content.\n## Level 2\nL2 content.\n### Level 3\nL3 content.\n#### Level 4\nL4 content.\n",
				"## Level 2\nL2 content.\n### Level 3\nL3 content.\n#### Level 4\nL4 content.\n",
				"### Level 3\nL3 content.\n#### Level 4\nL4 content.\n",
				"#### Level 4\nL4 content.\n",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			splitter := NewMdAstSplitter()
			ctx := context.Background()

			// Create input channel
			in := make(chan string, 1)
			in <- tt.input
			close(in)

			// Run the splitter
			out := splitter.Run(ctx, in)

			// Collect results
			var results []string
			for result := range out {
				results = append(results, result)
			}

			// Verify the number of outputs
			if len(results) != len(tt.expected) {
				t.Errorf("Expected %d outputs, got %d", len(tt.expected), len(results))
				t.Logf("Expected: %v", tt.expected)
				t.Logf("Got: %v", results)
				return
			}

			// Verify each output
			for i, expected := range tt.expected {
				if results[i] != expected {
					t.Errorf("Output %d mismatch.\nExpected: %q\nGot: %q", i, expected, results[i])
				}
			}
		})
	}
}

func TestMdAstSplitter_MultipleDocuments(t *testing.T) {
	splitter := NewMdAstSplitter()
	ctx := context.Background()

	// Create input channel with multiple documents
	in := make(chan string, 3)
	in <- "# Doc 1"
	in <- "# Doc 2\n\nParagraph.\n\n## Subtitle\n\nMore content."
	in <- "# Doc 3"
	close(in)

	// Run the splitter
	out := splitter.Run(ctx, in)

	// Collect results
	var results []string
	for result := range out {
		results = append(results, result)
	}

	expected := []string{
		"# Doc 1\n",
		"# Doc 2\nParagraph.\n## Subtitle\nMore content.\n",
		"## Subtitle\nMore content.\n",
		"# Doc 3\n",
	}

	if len(results) != len(expected) {
		t.Errorf("Expected %d outputs, got %d", len(expected), len(results))
		t.Logf("Expected: %v", expected)
		t.Logf("Got: %v", results)
		return
	}

	for i, exp := range expected {
		if results[i] != exp {
			t.Errorf("Output %d mismatch.\nExpected: %q\nGot: %q", i, exp, results[i])
		}
	}
}

func TestMdAstSplitter_ContextCancellation(t *testing.T) {
	splitter := NewMdAstSplitter()
	ctx, cancel := context.WithCancel(context.Background())

	// Create input channel with multiple documents
	in := make(chan string)
	go func() {
		defer close(in)
		for range 1000 {
			select {
			case in <- "# Heading":
			case <-ctx.Done():
				return
			}
			time.Sleep(1 * time.Millisecond)
		}
	}()

	// Run the splitter
	out := splitter.Run(ctx, in)

	// Cancel context after receiving a few results
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	// Collect results (should stop when context is cancelled)
	count := 0
	for range out {
		count++
	}

	// We should have received fewer than 1000 results due to cancellation
	if count >= 1000 {
		t.Errorf("Expected fewer than 1000 results due to cancellation, got %d", count)
	}
	
	t.Logf("Received %d results before cancellation", count)
}

func TestMdAstSplitter_EmptyInput(t *testing.T) {
	splitter := NewMdAstSplitter()
	ctx := context.Background()

	// Create empty input channel
	in := make(chan string)
	close(in)

	// Run the splitter
	out := splitter.Run(ctx, in)

	// Collect results
	var results []string
	for result := range out {
		results = append(results, result)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 outputs for empty input, got %d", len(results))
	}
}
