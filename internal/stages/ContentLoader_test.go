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

func TestContentLoader_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
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

	assert.Len(t, results, 0, "Should not return content on HTTP error")
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

func TestContentLoader_Concurrency(t *testing.T) {
	requestTimes := make(chan time.Time, 5)
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestTimes <- time.Now()
		time.Sleep(50 * time.Millisecond)
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><body><main>Content</main></body></html>"))
	}))
	defer server.Close()

	loader := NewContentLoader(3) // Allow 3 concurrent requests
	ctx := context.Background()

	in := make(chan string, 5)
	for i := 0; i < 5; i++ {
		in <- server.URL
	}
	close(in)

	start := time.Now()
	out := loader.Run(ctx, in)

	var results []string
	for result := range out {
		results = append(results, result)
	}
	elapsed := time.Since(start)

	assert.Len(t, results, 5)
	
	assert.Less(t, elapsed, 200*time.Millisecond, "Should complete faster with concurrency")
}

func TestContentLoader_DefaultConcurrency(t *testing.T) {
	loader := NewContentLoader(0)
	assert.Equal(t, 1, loader.concurrency, "Should default to 1 when concurrency is 0")

	loader = NewContentLoader(-5)
	assert.Equal(t, 1, loader.concurrency, "Should default to 1 when concurrency is negative")
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
	assert.Contains(t, content, "Text with extra spaces")
	assert.Contains(t, content, "Text with newlines")
}
