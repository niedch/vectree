package stages

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
)

type WebCrawler struct {
	sourceName  string
	url         string
	maxDepth    int
	selector    string
	concurrency int
	converter   *md.Converter
}

func NewWebCrawler(sourceName, url string, maxDepth int, selector string, concurrency int) *WebCrawler {
	if maxDepth <= 0 {
		maxDepth = 1
	}

	if concurrency <= 0 {
		concurrency = 1
	}

	sel := selector
	if sel == "" {
		sel = "main"
	}

	converter := md.NewConverter("", true, nil)
	return &WebCrawler{
		sourceName:  sourceName,
		url:         url,
		maxDepth:    maxDepth,
		selector:    sel,
		concurrency: concurrency,
		converter:   converter,
	}
}

type crawlPageResult struct {
	content string
	links   []string
}

func (w *WebCrawler) Run(ctx context.Context, _ <-chan any) <-chan string {
	out := make(chan string)

	go func() {
		defer close(out)

		visited := make(map[string]bool)
		currentURLs := []string{w.url}
		currentDepth := 0

		for currentDepth <= w.maxDepth && len(currentURLs) > 0 {
			var batch []string
			for _, u := range currentURLs {
				if !visited[u] {
					visited[u] = true
					batch = append(batch, u)
				}
			}

			if len(batch) == 0 {
				break
			}

			log.Printf("[%s] Crawling %d URLs at depth %d/%d", w.sourceName, len(batch), currentDepth, w.maxDepth)

			in := make(chan string, len(batch))
			for _, u := range batch {
				in <- u
			}
			close(in)

			var nextBatch []string
			var mu sync.Mutex

			results := WorkerPoolStage(ctx, in, w.concurrency, func(ctx context.Context, url string, results chan<- crawlPageResult) error {
				res, err := w.crawlPage(ctx, url)
				if err != nil {
					log.Printf("[%s] Error crawling %s: %v", w.sourceName, url, err)
					return err
				}
				results <- res
				return nil
			})

			for res := range results {
				if res.content != "" {
					select {
					case out <- res.content:
					case <-ctx.Done():
						return
					}
				}
				mu.Lock()
				for _, link := range res.links {
					if !visited[link] {
						nextBatch = append(nextBatch, link)
					}
				}
				mu.Unlock()
			}

			currentURLs = nextBatch
			currentDepth++
		}

		log.Printf("[%s] Crawling complete", w.sourceName)
	}()

	return out
}

func (w *WebCrawler) crawlPage(ctx context.Context, pageURL string) (crawlPageResult, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", pageURL, nil)
	if err != nil {
		return crawlPageResult{}, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return crawlPageResult{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return crawlPageResult{}, nil
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return crawlPageResult{}, err
	}

	sel := doc.Find(w.selector).First()
	var content string
	if sel.Length() > 0 {
		htmlContent, err := sel.Html()
		if err == nil && htmlContent != "" {
			markdown, err := w.converter.ConvertString(htmlContent)
			if err == nil {
				content = strings.TrimSpace(markdown)
			}
		}
	}

	if content == "" {
		log.Printf("[%s] No content found for selector %q in %s", w.sourceName, w.selector, pageURL)
	}

	baseURL, _ := url.Parse(pageURL)
	var links []string
	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists || href == "" || strings.HasPrefix(href, "#") {
			return
		}
		absoluteURL := resolveURL(baseURL, href)
		if absoluteURL != "" && isSameDomain(pageURL, absoluteURL) {
			links = append(links, absoluteURL)
		}
	})

	return crawlPageResult{content: content, links: links}, nil
}

func resolveURL(base *url.URL, href string) string {
	if href == "" {
		return ""
	}

	if strings.HasPrefix(href, "//") {
		href = base.Scheme + ":" + href
	}

	ref, err := url.Parse(href)
	if err != nil {
		return ""
	}

	resolved := base.ResolveReference(ref)
	resolved.Fragment = ""
	return resolved.String()
}

func isSameDomain(url1, url2 string) bool {
	u1, err := url.Parse(url1)
	if err != nil {
		return false
	}
	u2, err := url.Parse(url2)
	if err != nil {
		return false
	}
	return u1.Host == u2.Host
}
