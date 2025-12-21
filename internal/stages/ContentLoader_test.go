package stages

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestContentLoader(t *testing.T) {
	testHTML := `
<!DOCTYPE html>
<html>
<head><title>Test Page</title></head>
<body>
	<header>Header content</header>
	<main>
		<h1>Main Title</h1>
		<p>This is the main content.</p>
		<p>Another paragraph.</p>
	</main>
	<footer>Footer content</footer>
</body>
</html>
`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testHTML))
	}))
	defer server.Close()

	loader := NewContentLoader(1)
	ctx := context.Background()

	in := make(chan string, 1)
	in <- server.URL
	close(in)

	out := loader.Run(ctx, in)

	var results []string
	for result := range out {
		results = append(results, result)
	}

	assert.Len(t, results, 1)
	assert.Contains(t, results[0], "Main Title")
	assert.Contains(t, results[0], "This is the main content.")
	assert.Contains(t, results[0], "Another paragraph.")
	assert.NotContains(t, results[0], "Header content")
	assert.NotContains(t, results[0], "Footer content")
}

func TestContentLoader_NoMainTag(t *testing.T) {
	testHTML := `
<!DOCTYPE html>
<html>
<body>
	<div>Some content without main tag</div>
</body>
</html>
`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testHTML))
	}))
	defer server.Close()

	loader := NewContentLoader(1)
	ctx := context.Background()

	in := make(chan string, 1)
	in <- server.URL
	close(in)

	out := loader.Run(ctx, in)

	var results []string
	for result := range out {
		results = append(results, result)
	}

	assert.Len(t, results, 0, "Should not return content when no main tag found")
}

func TestContentLoader_MultipleURLs(t *testing.T) {
	testHTML1 := `<html><body><main>Content 1</main></body></html>`
	testHTML2 := `<html><body><main>Content 2</main></body></html>`
	testHTML3 := `<html><body><main>Content 3</main></body></html>`

	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)

		switch r.URL.Path {
		case "/page1":
			w.Write([]byte(testHTML1))
		case "/page2":
			w.Write([]byte(testHTML2))
		case "/page3":
			w.Write([]byte(testHTML3))
		}
	}))
	defer server.Close()

	loader := NewContentLoader(2)
	ctx := context.Background()

	in := make(chan string, 3)
	in <- server.URL + "/page1"
	in <- server.URL + "/page2"
	in <- server.URL + "/page3"
	close(in)

	out := loader.Run(ctx, in)

	var results []string
	for result := range out {
		results = append(results, result)
	}

	assert.Len(t, results, 3)
	assert.Equal(t, 3, callCount)

	contents := make(map[string]bool)
	for _, result := range results {
		contents[result] = true
	}

	assert.True(t, contents["Content 1"])
	assert.True(t, contents["Content 2"])
	assert.True(t, contents["Content 3"])
}

func TestContentLoader_NestedMainContent(t *testing.T) {
	testHTML := `
<!DOCTYPE html>
<html>
<body>
	<main>
		<article>
			<h1>Title</h1>
			<section>
				<h2>Section 1</h2>
				<p>Paragraph 1</p>
			</section>
			<section>
				<h2>Section 2</h2>
				<p>Paragraph 2</p>
			</section>
		</article>
	</main>
</body>
</html>
`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testHTML))
	}))
	defer server.Close()

	loader := NewContentLoader(1)
	ctx := context.Background()

	in := make(chan string, 1)
	in <- server.URL
	close(in)

	out := loader.Run(ctx, in)

	var results []string
	for result := range out {
		results = append(results, result)
	}

	assert.Len(t, results, 1)
	assert.Contains(t, results[0], "Title")
	assert.Contains(t, results[0], "Section 1")
	assert.Contains(t, results[0], "Paragraph 1")
	assert.Contains(t, results[0], "Section 2")
	assert.Contains(t, results[0], "Paragraph 2")
}

func TestContentLoader_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><body><main>Content</main></body></html>"))
	}))
	defer server.Close()

	loader := NewContentLoader(1)
	ctx, cancel := context.WithCancel(context.Background())

	in := make(chan string, 1)
	in <- server.URL
	close(in)

	out := loader.Run(ctx, in)

	cancel()

	var results []string
	for result := range out {
		results = append(results, result)
	}

	assert.True(t, len(results) <= 1)
}

func TestContentLoader_WhitespaceHandling(t *testing.T) {
	testHTML := `
<!DOCTYPE html>
<html>
<body>
	<main>
		<p>   Text   with   extra   spaces   </p>
		
		<p>
			Text
			with
			newlines
		</p>
	</main>
</body>
</html>
`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testHTML))
	}))
	defer server.Close()

	loader := NewContentLoader(1)
	ctx := context.Background()

	in := make(chan string, 1)
	in <- server.URL
	close(in)

	out := loader.Run(ctx, in)

	var results []string
	for result := range out {
		results = append(results, result)
	}

	assert.Len(t, results, 1)

	content := results[0]
	// The library normalizes multiple spaces to single spaces
	assert.Contains(t, content, "Text with extra spaces")
	// The library preserves newlines in the source HTML
	assert.Contains(t, content, "Text")
	assert.Contains(t, content, "with")
	assert.Contains(t, content, "newlines")
}

func TestContentLoader_OutputIsMarkdown(t *testing.T) {
	testHTML := `
<!DOCTYPE html>
<html>
<body>
	<main>
		<h1>Main Heading</h1>
		<h2>Subheading</h2>
		<p>This is a <strong>bold</strong> and <em>italic</em> text.</p>
		<ul>
			<li>List item 1</li>
			<li>List item 2</li>
		</ul>
		<a href="https://example.com">Example Link</a>
		<code>inline code</code>
		<pre><code>code block</code></pre>
	</main>
</body>
</html>
`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testHTML))
	}))
	defer server.Close()

	loader := NewContentLoader(1)
	ctx := context.Background()

	in := make(chan string, 1)
	in <- server.URL
	close(in)

	out := loader.Run(ctx, in)

	var results []string
	for result := range out {
		results = append(results, result)
	}

	assert.Len(t, results, 1)
	markdown := results[0]

	// Verify markdown syntax for headers
	assert.Contains(t, markdown, "# Main Heading", "Should contain markdown H1 header")
	assert.Contains(t, markdown, "## Subheading", "Should contain markdown H2 header")

	// Verify markdown syntax for bold and italic
	assert.Contains(t, markdown, "**bold**", "Should contain markdown bold syntax")
	assert.Contains(t, markdown, "_italic_", "Should contain markdown italic syntax")

	// Verify markdown syntax for lists
	assert.Contains(t, markdown, "- List item 1", "Should contain markdown list syntax")
	assert.Contains(t, markdown, "- List item 2", "Should contain markdown list syntax")

	// Verify markdown syntax for links
	assert.Contains(t, markdown, "[Example Link](https://example.com)", "Should contain markdown link syntax")

	// Verify markdown syntax for code
	assert.Contains(t, markdown, "`inline code`", "Should contain markdown inline code syntax")
	assert.Contains(t, markdown, "```", "Should contain markdown code block syntax")

	// Verify it's NOT HTML
	assert.NotContains(t, markdown, "<h1>", "Should not contain HTML tags")
	assert.NotContains(t, markdown, "<p>", "Should not contain HTML tags")
	assert.NotContains(t, markdown, "<strong>", "Should not contain HTML tags")
	assert.NotContains(t, markdown, "<ul>", "Should not contain HTML tags")
	assert.NotContains(t, markdown, "<li>", "Should not contain HTML tags")
}
