package stages

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMdAstSplitter_Run(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []SectionWithLevel
	}{
		{
			name:  "Single heading",
			input: "# Hello World",
			expected: []SectionWithLevel{
				{Text: "# Hello World\n", Level: 1},
			},
		},
		{
			name: "Heading with paragraph",
			input: `# Title

This is a paragraph.`,
			expected: []SectionWithLevel{
				{Text: "# Title\nThis is a paragraph.\n", Level: 1},
			},
		},
		{
			name: "Multiple top-level headings",
			input: `# Title One

Paragraph one.

# Title Two

Paragraph two.`,
			expected: []SectionWithLevel{
				{Text: "# Title One\nParagraph one.\n", Level: 1},
				{Text: "# Title Two\nParagraph two.\n", Level: 1},
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
			expected: []SectionWithLevel{
				{Text: "# Main Title\nIntroduction paragraph.\n## Section One\nSection one content.\n### Subsection 1.1\nSubsection content.\n## Section Two\nSection two content.\n", Level: 1},
				{Text: "## Section One\nSection one content.\n### Subsection 1.1\nSubsection content.\n", Level: 2},
				{Text: "### Subsection 1.1\nSubsection content.\n", Level: 3},
				{Text: "## Section Two\nSection two content.\n", Level: 2},
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
			expected: []SectionWithLevel{
				{Text: "# First Title\nFirst intro.\n## First Section\nFirst section content.\n", Level: 1},
				{Text: "## First Section\nFirst section content.\n", Level: 2},
				{Text: "# Second Title\nSecond intro.\n## Second Section\nSecond section content.\n", Level: 1},
				{Text: "## Second Section\nSecond section content.\n", Level: 2},
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
			expected: []SectionWithLevel{
				{Text: "# Main Title\nFirst paragraph.\n## Section One\nContent for section one.\n### Subsection 1.1\nDetails here.\n### Subsection 1.2\nMore details.\n## Section Two\nContent for section two.\n", Level: 1},
				{Text: "## Section One\nContent for section one.\n### Subsection 1.1\nDetails here.\n### Subsection 1.2\nMore details.\n", Level: 2},
				{Text: "### Subsection 1.1\nDetails here.\n", Level: 3},
				{Text: "### Subsection 1.2\nMore details.\n", Level: 3},
				{Text: "## Section Two\nContent for section two.\n", Level: 2},
			},
		},
		{
			name:     "No headings",
			input:    "Just a paragraph with no headings.",
			expected: []SectionWithLevel{},
		},
		{
			name: "Multiple level 2 headings",
			input: `## Section One

Content one.

## Section Two

Content two.

## Section Three

Content three.`,
			expected: []SectionWithLevel{
				{Text: "## Section One\nContent one.\n", Level: 2},
				{Text: "## Section Two\nContent two.\n", Level: 2},
				{Text: "## Section Three\nContent three.\n", Level: 2},
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
			expected: []SectionWithLevel{
				{Text: "# Level 1\nL1 content.\n## Level 2\nL2 content.\n### Level 3\nL3 content.\n#### Level 4\nL4 content.\n", Level: 1},
				{Text: "## Level 2\nL2 content.\n### Level 3\nL3 content.\n#### Level 4\nL4 content.\n", Level: 2},
				{Text: "### Level 3\nL3 content.\n#### Level 4\nL4 content.\n", Level: 3},
				{Text: "#### Level 4\nL4 content.\n", Level: 4},
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
			var results []SectionWithLevel
			for result := range out {
				results = append(results, result)
			}

			// Verify the number of outputs
			assert.Equal(t, len(tt.expected), len(results), "Number of outputs mismatch")

			// Verify each output
			for i, expected := range tt.expected {
				assert.Equal(t, expected.Text, results[i].Text, "Output %d text mismatch", i)
				assert.Equal(t, expected.Level, results[i].Level, "Output %d level mismatch", i)
				// DocumentId should be set (non-empty)
				assert.NotEmpty(t, results[i].DocumentId, "Output %d DocumentId should not be empty", i)
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
	var results []SectionWithLevel
	for result := range out {
		results = append(results, result)
	}

	expected := []SectionWithLevel{
		{Text: "# Doc 1\n", Level: 1},
		{Text: "# Doc 2\nParagraph.\n## Subtitle\nMore content.\n", Level: 1},
		{Text: "## Subtitle\nMore content.\n", Level: 2},
		{Text: "# Doc 3\n", Level: 1},
	}

	assert.Equal(t, len(expected), len(results), "Number of outputs mismatch")

	// Sort by text to account for parallel processing order
	sort.Slice(results, func(i, j int) bool {
		return results[i].Text < results[j].Text
	})
	sort.Slice(expected, func(i, j int) bool {
		return expected[i].Text < expected[j].Text
	})

	// Track document IDs to verify they're different for different documents
	docIds := make(map[string]bool)
	
	for i, exp := range expected {
		assert.Equal(t, exp.Text, results[i].Text, "Output %d text mismatch", i)
		assert.Equal(t, exp.Level, results[i].Level, "Output %d level mismatch", i)
		// DocumentId should be set (non-empty)
		assert.NotEmpty(t, results[i].DocumentId, "Output %d DocumentId should not be empty", i)
		docIds[results[i].DocumentId] = true
	}
	
	// We should have 3 different document IDs (one for each input document)
	assert.Equal(t, 3, len(docIds), "Expected 3 different document IDs")
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
	assert.Less(t, count, 1000, "Expected fewer than 1000 results due to cancellation")
	
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
	var results []SectionWithLevel
	for result := range out {
		results = append(results, result)
	}

	assert.Empty(t, results, "Expected 0 outputs for empty input")
}
