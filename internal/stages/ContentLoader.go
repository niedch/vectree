package stages

import (
	"context"
	"log"
	"net/http"
	"slices"
	"strings"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"golang.org/x/net/html"
)

type ContentLoader struct {
	concurrency int
	converter   *md.Converter
}

func NewContentLoader(concurrency int) *ContentLoader {
	if concurrency <= 0 {
		concurrency = 1
	}
	converter := md.NewConverter("", true, nil)
	return &ContentLoader{
		concurrency: concurrency,
		converter:   converter,
	}
}

func (l ContentLoader) Run(ctx context.Context, in <-chan string) <-chan string {
	out := make(chan string)

	go func() {
		defer close(out)

		urlCount := 0
		var totalSize uint64 = 0

		innerOut := WorkerPoolStage(ctx, in, l.concurrency, func(ctx context.Context, url string, out chan<- string) error {
			content, err := l.fetchAndExtractMain(ctx, url)
			if err != nil {
				return err
			}

			if content == "" {
				return nil
			}

			out <- content
			return nil
		})

		for content := range innerOut {
			urlCount++
			totalSize += uint64(len(content))

			select {
			case out <- content:
			case <-ctx.Done():
				return
			}
		}

		log.Printf("ContentLoader: Loaded %d URLs, total size: %d bytes (%.2f MB)\n",
			urlCount, totalSize, float64(totalSize)/(1024*1024))
	}()

	return out
}

func (l ContentLoader) fetchAndExtractMain(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.Printf("Error creating request for %s: %v", url, err)
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Error fetching %s: %v", url, err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Error: received status code %d for %s", resp.StatusCode, url)
		return "", err
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		log.Printf("Error parsing HTML for %s: %v", url, err)
		return "", err
	}

	mainContent := l.extractMainContent(doc)

	if mainContent == "" {
		log.Printf("Warning: No <main> tag found in %s", url)
	}

	return mainContent, nil
}

func (l ContentLoader) extractMainContent(n *html.Node) string {
	mainNode := findMainTag(n)
	if mainNode == nil {
		return ""
	}

	return l.htmlNodeToMarkdown(mainNode)
}

func findMainTag(n *html.Node) *html.Node {
	if n.Type == html.ElementNode && (n.Data == "main" || hasClass(n, "main-content")) {
		return n
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if result := findMainTag(c); result != nil {
			return result
		}
	}

	return nil
}

func hasClass(n *html.Node, class string) bool {
	for _, attr := range n.Attr {
		if attr.Key != "class" {
			continue
		}

		if slices.Contains(strings.Fields(attr.Val), class) {
			return true
		}
	}

	return false
}

func (l ContentLoader) htmlNodeToMarkdown(n *html.Node) string {
	// Convert the html.Node to an HTML string
	var sb strings.Builder
	html.Render(&sb, n)
	htmlContent := sb.String()

	// Convert HTML to Markdown using the library
	markdown, err := l.converter.ConvertString(htmlContent)
	if err != nil {
		log.Printf("Error converting HTML to Markdown: %v", err)
		return ""
	}

	return strings.TrimSpace(markdown)
}
