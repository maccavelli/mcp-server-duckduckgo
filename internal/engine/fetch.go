package engine

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/maccavelli/mcp-server-duckduckgo/internal/config"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

// FetchedContent represents extracted page content from a URL.
type FetchedContent struct {
	URL         string `json:"url"`
	Title       string `json:"title"`
	Content     string `json:"content"`
	ContentType string `json:"content_type"`
	StatusCode  int    `json:"status_code"`
}

// noiseSelectors are CSS selectors for elements that pollute readable content.
var noiseSelectors = []string{
	"script", "style", "noscript", "iframe", "svg",
	"nav", "header", "footer",
	".nav", ".navbar", ".navigation", ".menu", ".sidebar",
	".header", ".footer", ".breadcrumb",
	".ad", ".ads", ".advertisement", ".sponsored",
	".cookie-banner", ".cookie-consent", ".gdpr",
	".social-share", ".share-buttons", ".social-links",
	".comments", ".comment-section",
	".related-posts", ".recommended", ".you-may-also-like",
	".popup", ".modal", ".overlay",
	"[role='navigation']", "[role='banner']", "[role='contentinfo']",
	".sr-only", "[aria-hidden='true']",
}

// contentSelectors are CSS selectors tried in priority order to find the
// main article body. The first match wins.
var contentSelectors = []string{
	htmlTagArticle,
	"[role='main']",
	"main",
	".post-content",
	".article-content",
	".entry-content",
	".content-body",
	".story-body",
	".article-body",
	".post-body",
	".page-content",
	"#content",
	".content",
}

// FetchContent retrieves a URL using the hardened HTTP client and extracts
// clean, readable text content by stripping navigation, ads, and scripts.
func (e *SearchEngine) FetchContent(parentCtx context.Context, rawURL string) (*FetchedContent, error) {
	ctx, cancel := context.WithTimeout(parentCtx, 15*time.Second)
	defer cancel()

	var resp *http.Response
	var fetchErr error
	backoff := 1 * time.Second

	for attempt := range 3 {
		req, err := e.newRequest(ctx, http.MethodGet, rawURL, http.NoBody)
		if err != nil {
			return nil, fmt.Errorf("failed to create fetch request: %w", err)
		}
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
		req.Header.Set("Accept-Language", "en-US,en;q=0.9")

		resp, fetchErr = e.Client.Do(req)
		if fetchErr == nil {
			if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusServiceUnavailable {
				slog.Warn("fetch failed with rate limit or service unavailable, backing off", "url", rawURL, "status", resp.StatusCode, "attempt", attempt+1)
				select {
				case <-ctx.Done():
					closeResponseBody(resp.Body)
					return nil, ctx.Err()
				case <-time.After(backoff):
					backoff *= 2
					if attempt < 2 {
						closeResponseBody(resp.Body)
					}
					continue
				}
			}
			break
		}

		slog.Warn("fetch failed, backing off", "url", rawURL, "error", fetchErr, "attempt", attempt+1)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(backoff):
			backoff *= 2
		}
	}

	if fetchErr != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", fetchErr)
	}
	defer closeResponseBody(resp.Body)

	// Detect CAPTCHA/Challenge and fallback to Playwright (Workaround 2)
	bodyPreview, err := io.ReadAll(io.LimitReader(resp.Body, 32768))
	if err != nil {
		return nil, fmt.Errorf("failed to read response preview: %w", err)
	}
	resp.Body = io.NopCloser(io.MultiReader(bytes.NewReader(bodyPreview), resp.Body))
	challenge := detectDDGChallenge(bodyPreview)

	if resp.StatusCode == http.StatusForbidden || challenge != "" {
		slog.Warn("fetch detected CAPTCHA/403, falling back to playwright", "url", rawURL)
		pwBytes, pwErr := e.fetchWithPlaywright(ctx, rawURL)
		if pwErr == nil {
			resp.StatusCode = http.StatusOK
			resp.Body = io.NopCloser(bytes.NewReader(pwBytes))
			resp.Header.Set("Content-Type", "text/html")
		} else {
			slog.Error("playwright fallback failed", "url", rawURL, "error", pwErr)
		}
	}

	ct := resp.Header.Get("Content-Type")
	result := &FetchedContent{
		URL:         rawURL,
		StatusCode:  resp.StatusCode,
		ContentType: ct,
	}

	// Non-200 response: return status info without content.
	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("fetch failed with status %d", resp.StatusCode)
	}

	// Binary content (images, PDFs, etc.): return metadata only.
	if !strings.Contains(ct, "text/html") && !strings.Contains(ct, "text/plain") &&
		!strings.Contains(ct, "application/xhtml") {
		result.Content = fmt.Sprintf("[Binary content: %s]", ct)
		return result, nil
	}

	// Plain text: return as-is.
	if strings.Contains(ct, "text/plain") {
		body, err := io.ReadAll(io.LimitReader(resp.Body, config.MaxBodyBytes))
		if err != nil {
			return nil, fmt.Errorf("failed to read plain text body: %w", err)
		}
		result.Content = string(body)
		return result, nil
	}

	// HTML: parse and extract readable content.
	doc, err := goquery.NewDocumentFromReader(io.LimitReader(resp.Body, config.MaxBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Extract title.
	result.Title = strings.TrimSpace(doc.Find("title").First().Text())

	// Remove noise elements.
	for _, sel := range noiseSelectors {
		doc.Find(sel).Remove()
	}

	// Try to find the main content container.
	var contentNode *goquery.Selection
	for _, sel := range contentSelectors {
		found := doc.Find(sel)
		if found.Length() > 0 {
			contentNode = found.First()
			break
		}
	}

	// Fallback to heuristic text density if no content container found.
	if contentNode == nil {
		var bestNode *goquery.Selection
		maxScore := 0
		doc.Find("div, section, main, article").Each(func(i int, s *goquery.Selection) {
			textLen := len(strings.TrimSpace(s.Text()))
			pCount := s.Find("p").Length()
			score := textLen + (pCount * 100)
			if score > maxScore {
				maxScore = score
				bestNode = s
			}
		})
		if bestNode != nil && maxScore > 100 {
			contentNode = bestNode
		} else {
			contentNode = doc.Find("body")
		}
	}

	if contentNode == nil || contentNode.Length() == 0 {
		result.Content = "[No readable content extracted]"
		return result, nil
	}

	// Extract text with structure preservation.
	result.Content = extractStructuredText(contentNode)

	// Truncate if excessively long (protect token budgets).
	const maxContentRunes = 15000
	runes := []rune(result.Content)
	if len(runes) > maxContentRunes {
		result.Content = string(runes[:maxContentRunes]) + "\n\n[Content truncated at 15,000 characters]"
	}

	slog.Debug("content fetched successfully", "url", rawURL, "title", result.Title, "content_len", len(result.Content))
	return result, nil
}

// extractStructuredText walks the DOM and produces clean text with basic
// structure markers (headings, paragraphs, list items).
func extractStructuredText(s *goquery.Selection) string {
	var b strings.Builder

	var walkDOM func(n *html.Node, depth int, inPre bool)
	walkDOM = func(n *html.Node, depth int, inPre bool) {
		if depth > 1000 {
			return
		}

		if n.Type == html.TextNode {
			b.WriteString(n.Data)
			return
		}

		if n.Type == html.ElementNode {
			tag := n.Data
			// Intentionally skip a tags' href extraction to prevent token bloat
			switch tag {
			case "pre":
				inPre = true
				b.WriteString("\n```\n")
			case "h1", "h2", "h3", "h4", "h5", "h6":
				b.WriteString("\n## ")
			case "li":
				b.WriteString("\n- ")
			case "blockquote":
				b.WriteString("\n> ")
			case "p", "div", "br", htmlTagArticle, "section":
				b.WriteString("\n")
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walkDOM(c, depth+1, inPre)
		}

		if n.Type == html.ElementNode {
			tag := n.Data
			switch tag {
			case "pre":
				b.WriteString("\n```\n")
			case "td", "th":
				b.WriteString(" | ")
			case "tr":
				b.WriteString("\n")
			case "p", "div", htmlTagArticle, "section":
				b.WriteString("\n")
			}
		}
	}

	for _, node := range s.Nodes {
		walkDOM(node, 0, false)
	}

	// O(N) cleanup using single-pass regex for spaces outside of pre blocks
	parts := strings.Split(b.String(), "```")
	spaceRe := regexp.MustCompile(`[ \t]+`)
	for i := range parts {
		if i%2 == 0 { // non-pre blocks
			parts[i] = spaceRe.ReplaceAllString(parts[i], " ")
		}
	}
	result := strings.Join(parts, "```")

	// Collapse multiple blank lines.
	for strings.Contains(result, "\n\n\n") {
		result = strings.ReplaceAll(result, "\n\n\n", "\n\n")
	}
	return strings.TrimSpace(result)
}
