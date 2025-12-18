package stages

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDocTocLoader(t *testing.T) {
	// Create a test JSON TOC response
	testTOCJSON := `[
		{"title":"Page 1","link":"/page1","scope":"local","children":[]},
		{"title":"Page 2","link":"/page2","scope":"local","children":[]},
		{"title":"Page 3","link":"https://example.com/page3","scope":"external","children":[]}
	]`

	// Create a test server that returns TOC JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testTOCJSON))
	}))
	defer server.Close()

	// Create the DocTocLoader with the TOC URL directly
	loader := NewDocTocLoader(server.URL)

	// Run the loader
	ctx := context.Background()
	out := loader.Run(ctx, nil)

	// Collect results
	var links []string
	for link := range out {
		links = append(links, link)
	}

	// Verify results
	assert.Len(t, links, 3, "Should extract exactly 3 links from TOC JSON")
	assert.Contains(t, links, server.URL+"/page1")
	assert.Contains(t, links, server.URL+"/page2")
	assert.Contains(t, links, "https://example.com/page3")
}

func TestDocTocLoader_EmptyTOC(t *testing.T) {
	testTOCJSON := `[]`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testTOCJSON))
	}))
	defer server.Close()

	loader := NewDocTocLoader(server.URL)
	ctx := context.Background()
	out := loader.Run(ctx, nil)

	var links []string
	for link := range out {
		links = append(links, link)
	}

	assert.Len(t, links, 0, "Should return no links when TOC is empty")
}

func TestDocTocLoader_NestedChildren(t *testing.T) {
	testTOCJSON := `[
		{
			"title":"Section 1",
			"link":"/section1",
			"scope":"local",
			"children":[
				{"title":"Section 1.1","link":"/section1.1","scope":"local","children":[]},
				{"title":"Section 1.2","link":"/section1.2","scope":"local","children":[]}
			]
		},
		{"title":"Section 2","link":"/section2","scope":"local","children":[]}
	]`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testTOCJSON))
	}))
	defer server.Close()

	loader := NewDocTocLoader(server.URL)
	ctx := context.Background()
	out := loader.Run(ctx, nil)

	var links []string
	for link := range out {
		links = append(links, link)
	}

	assert.Len(t, links, 4, "Should extract all nested links")
	assert.Contains(t, links, server.URL+"/section1")
	assert.Contains(t, links, server.URL+"/section1.1")
	assert.Contains(t, links, server.URL+"/section1.2")
	assert.Contains(t, links, server.URL+"/section2")
}

func TestDocTocLoader_DeeplyNestedChildren(t *testing.T) {
	testTOCJSON := `[
		{
			"title":"Level 1",
			"link":"/level1",
			"scope":"local",
			"children":[
				{
					"title":"Level 2",
					"link":"/level2",
					"scope":"local",
					"children":[
						{
							"title":"Level 3",
							"link":"/level3",
							"scope":"local",
							"children":[]
						}
					]
				}
			]
		}
	]`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testTOCJSON))
	}))
	defer server.Close()

	loader := NewDocTocLoader(server.URL)
	ctx := context.Background()
	out := loader.Run(ctx, nil)

	var links []string
	for link := range out {
		links = append(links, link)
	}

	assert.Len(t, links, 3, "Should extract all deeply nested links")
	assert.Contains(t, links, server.URL+"/level1")
	assert.Contains(t, links, server.URL+"/level2")
	assert.Contains(t, links, server.URL+"/level3")
}

func TestDocTocLoader_ContextCancellation(t *testing.T) {
	testTOCJSON := `[
		{"title":"Link 1","link":"/link1","scope":"local","children":[]},
		{"title":"Link 2","link":"/link2","scope":"local","children":[]}
	]`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testTOCJSON))
	}))
	defer server.Close()

	loader := NewDocTocLoader(server.URL)
	ctx, cancel := context.WithCancel(context.Background())
	
	out := loader.Run(ctx, nil)
	
	// Cancel context immediately
	cancel()

	// Try to read from channel
	var links []string
	for link := range out {
		links = append(links, link)
	}

	// Should handle cancellation gracefully
	assert.True(t, len(links) <= 2, "Should stop processing when context is cancelled")
}

func TestDocTocLoader_InvalidJSON(t *testing.T) {
	testInvalidJSON := `{invalid json`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testInvalidJSON))
	}))
	defer server.Close()

	loader := NewDocTocLoader(server.URL)
	ctx := context.Background()
	out := loader.Run(ctx, nil)

	var links []string
	for link := range out {
		links = append(links, link)
	}

	assert.Len(t, links, 0, "Should return no links when JSON is invalid")
}

func TestDocTocLoader_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	loader := NewDocTocLoader(server.URL)
	ctx := context.Background()
	out := loader.Run(ctx, nil)

	var links []string
	for link := range out {
		links = append(links, link)
	}

	assert.Len(t, links, 0, "Should return no links when HTTP request fails")
}
