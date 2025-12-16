package stages

import (
	"context"
	"log"
	"net/http"
	"strings"

	"golang.org/x/net/html"
)

type ContentLoader struct {
	concurrency int
}

func NewContentLoader(concurrency int) *ContentLoader {
	if concurrency <= 0 {
		concurrency = 1
	}
	return &ContentLoader{
		concurrency: concurrency,
	}
}

func (l ContentLoader) Run(ctx context.Context, in <-chan string) <-chan string {
	out := make(chan string)

	go func() {
		defer close(out)

		sem := make(chan struct{}, l.concurrency)

		done := make(chan struct{})
		activeCount := 0

		for url := range in {
			activeCount++
			
			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				return
			}

			go func(url string) {
				defer func() {
					<-sem
					done <- struct{}{}
				}()

				content, err := l.fetchAndExtractMain(ctx, url)
				
				if err != nil {
					return
				}

				if content == "" {
					return
				}

				select {
				case out <- content:
				case <-ctx.Done():
					return
				}
			}(url)
		}

		for i := 0; i < activeCount; i++ {
			<-done
		}
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

	mainContent := extractMainContent(doc)
	
	if mainContent == "" {
		log.Printf("Warning: No <main> tag found in %s", url)
	}

	return mainContent, nil
}

func extractMainContent(n *html.Node) string {
	mainNode := findMainTag(n)
	if mainNode == nil {
		return ""
	}

	return extractText(mainNode)
}

func findMainTag(n *html.Node) *html.Node {
	if n.Type == html.ElementNode && n.Data == "main" {
		return n
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if result := findMainTag(c); result != nil {
			return result
		}
	}

	return nil
}

func extractText(n *html.Node) string {
	var sb strings.Builder
	extractTextRecursive(n, &sb)
	
	// Normalize whitespace: replace multiple spaces/newlines with single space
	text := sb.String()
	text = strings.Join(strings.Fields(text), " ")
	
	return strings.TrimSpace(text)
}

func extractTextRecursive(n *html.Node, sb *strings.Builder) {
	if n.Type == html.TextNode {
		text := n.Data
		if text != "" {
			sb.WriteString(text)
			sb.WriteString(" ")
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		extractTextRecursive(c, sb)
	}
}
