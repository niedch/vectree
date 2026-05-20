package stages

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type DocTocLoader struct {
	url        string
	sourceName string
}

func NewDocTocLoader(sourceName string, url string) *DocTocLoader {
	return &DocTocLoader{
		sourceName: sourceName,
		url:        url,
	}
}

// TOCEntry represents a single entry in the table of contents
type TOCEntry struct {
	Title    string     `json:"title"`
	Link     string     `json:"link"`
	Scope    string     `json:"scope"`
	Children []TOCEntry `json:"children"`
}

func (l DocTocLoader) Run(ctx context.Context, _ <-chan any) <-chan string {
	out := make(chan string)

	go func() {
		defer close(out)

		log.Printf("[%s] Started loading TOC", l.sourceName)
		req, err := http.NewRequestWithContext(ctx, "GET", l.url, nil)
		if err != nil {
			log.Printf("Error creating request: %v", err)
			return
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("Error downloading TOC: %v", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("Error: received status code %d", resp.StatusCode)
			return
		}

		// Parse the JSON
		var tocEntries []TOCEntry
		if err := json.NewDecoder(resp.Body).Decode(&tocEntries); err != nil {
			log.Printf("Error parsing JSON: %v", err)
			return
		}

		// Extract all links recursively
		links := extractAllLinks(tocEntries)
		log.Printf("Extracted %d links from TOC", len(links))

		// Get the base URL for constructing absolute URLs
		baseURL, err := l.getBaseURL()
		if err != nil {
			log.Printf("Error getting base URL: %v", err)
			return
		}

		log.Printf("[%s] Loaded %d Links", l.sourceName, len(links))
		// Send links to the output channel
		for _, link := range links {
			// Convert relative links to absolute URLs
			absoluteURL := makeAbsoluteURL(baseURL, link)

			select {
			case out <- absoluteURL:
			case <-ctx.Done():
				return
			}
		}
	}()

	return out
}

func (l DocTocLoader) getBaseURL() (string, error) {
	u, err := url.Parse(l.url)
	if err != nil {
		return "", err
	}
	return u.Scheme + "://" + u.Host, nil
}

// extractAllLinks recursively extracts all links from TOC entries
func extractAllLinks(entries []TOCEntry) []string {
	var links []string

	for _, entry := range entries {
		if entry.Link != "" {
			links = append(links, entry.Link)
		}

		// Recursively extract links from children
		if len(entry.Children) > 0 {
			childLinks := extractAllLinks(entry.Children)
			links = append(links, childLinks...)
		}
	}

	return links
}

func makeAbsoluteURL(baseURL, link string) string {
	if strings.HasPrefix(link, "http://") || strings.HasPrefix(link, "https://") {
		return link
	}

	if strings.HasPrefix(link, "/") {
		return baseURL + link
	}

	return baseURL + "/" + link
}
