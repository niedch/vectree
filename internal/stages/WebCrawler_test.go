package stages

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWebCrawler_SinglePage(t *testing.T) {
	server := httptest.NewServer(http.FileServer(http.Dir("../../test_data/webcrawler")))
	defer server.Close()

	crawler := NewWebCrawler("test", server.URL+"/simple.html", 1, "main", 1)
	ctx := context.Background()
	out := crawler.Run(ctx, nil)

	var results []string
	for content := range out {
		results = append(results, content)
	}

	assert.Len(t, results, 1)
	assert.Contains(t, results[0], "# Hello")
	assert.Contains(t, results[0], "World")
}

func TestWebCrawler_DepthTwo(t *testing.T) {
	server := httptest.NewServer(http.FileServer(http.Dir("../../test_data/webcrawler")))
	defer server.Close()

	crawler := NewWebCrawler("test", server.URL+"/page1.html", 2, "main", 1)
	ctx := context.Background()
	out := crawler.Run(ctx, nil)

	var results []string
	for content := range out {
		results = append(results, content)
	}

	assert.Len(t, results, 2)
	assert.Contains(t, results[0], "Page 1")
	assert.Contains(t, results[1], "Page 2")
}

func TestWebCrawler_CustomSelector(t *testing.T) {
	server := httptest.NewServer(http.FileServer(http.Dir("../../test_data/webcrawler")))
	defer server.Close()

	crawler := NewWebCrawler("test", server.URL+"/custom_selector.html", 1, ".content", 1)
	ctx := context.Background()
	out := crawler.Run(ctx, nil)

	var results []string
	for content := range out {
		results = append(results, content)
	}

	assert.Len(t, results, 1)
	assert.Contains(t, results[0], "Title")
	assert.NotContains(t, results[0], "Sidebar")
}

func TestWebCrawler_NoMatch(t *testing.T) {
	server := httptest.NewServer(http.FileServer(http.Dir("../../test_data/webcrawler")))
	defer server.Close()

	crawler := NewWebCrawler("test", server.URL+"/no_match.html", 1, "main", 1)
	ctx := context.Background()
	out := crawler.Run(ctx, nil)

	var results []string
	for content := range out {
		results = append(results, content)
	}

	assert.Empty(t, results)
}

func TestWebCrawler_ExternalLinksIgnored(t *testing.T) {
	server := httptest.NewServer(http.FileServer(http.Dir("../../test_data/webcrawler")))
	defer server.Close()

	crawler := NewWebCrawler("test", server.URL+"/external_links.html", 2, "main", 1)
	ctx := context.Background()
	out := crawler.Run(ctx, nil)

	var results []string
	for content := range out {
		results = append(results, content)
	}

	assert.Len(t, results, 1)
	assert.Contains(t, results[0], "Home")
}

func TestWebCrawler_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.FileServer(http.Dir("../../test_data/webcrawler")))
	defer server.Close()

	crawler := NewWebCrawler("test", server.URL+"/simple.html", 1, "main", 1)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	out := crawler.Run(ctx, nil)
	var results []string
	for content := range out {
		results = append(results, content)
	}

	assert.Empty(t, results)
}

func TestWebCrawler_EmptyContent(t *testing.T) {
	server := httptest.NewServer(http.FileServer(http.Dir("../../test_data/webcrawler")))
	defer server.Close()

	crawler := NewWebCrawler("test", server.URL+"/empty_content.html", 1, "main", 1)
	ctx := context.Background()
	out := crawler.Run(ctx, nil)

	var results []string
	for content := range out {
		results = append(results, content)
	}

	assert.Empty(t, results)
}

func TestIsSameDomain(t *testing.T) {
	assert.True(t, isSameDomain("https://example.com/page1", "https://example.com/page2"))
	assert.False(t, isSameDomain("https://example.com", "https://other.com"))
	assert.True(t, isSameDomain("http://example.com:8080", "http://example.com:8080/foo"))
	assert.False(t, isSameDomain("https://example.com", ""))
	assert.False(t, isSameDomain("", "https://example.com"))
}

func TestResolveURL(t *testing.T) {
	base, _ := url.Parse("https://example.com/docs/")
	assert.Equal(t, "https://example.com/page", resolveURL(base, "/page"))
	assert.Equal(t, "https://example.com/docs/sub", resolveURL(base, "sub"))
	assert.Equal(t, "https://other.com", resolveURL(base, "https://other.com"))
	assert.Equal(t, "", resolveURL(base, ""))
}
