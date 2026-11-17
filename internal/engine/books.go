package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/maccavelli/mcp-server-duckduckgo/internal/config"
	"github.com/maccavelli/mcp-server-duckduckgo/internal/models"
)

const (
	openLibrarySearchURL = "https://openlibrary.org/search.json"
	openLibraryFields    = "key,title,subtitle,author_name,first_publish_year,number_of_pages_median,cover_i,edition_count"
	openLibraryUserAgent = "mcp-server-duckduckgo (+https://github.com/maccavelli/mcp-server-duckduckgo)"
	openLibraryMaxLimit  = 50
)

// olBareWorkKey matches a work id with no /works/ prefix. Author search emits
// that shape; /search.json emits /works/OL…W. A bare id is not a valid URL.
var olBareWorkKey = regexp.MustCompile(`^OL\d+W$`)

type olWorkDoc struct {
	Key              string   `json:"key"`
	Title            string   `json:"title"`
	Subtitle         string   `json:"subtitle"`
	AuthorName       []string `json:"author_name"`
	FirstPublishYear int      `json:"first_publish_year"`
	PagesMedian      float64  `json:"number_of_pages_median"`
	CoverI           int      `json:"cover_i"`
	EditionCount     int      `json:"edition_count"`
}

type olSearchResponse struct {
	Docs []olWorkDoc `json:"docs"`
}

// BookSearch looks up works in the Open Library Search API.
func (e *SearchEngine) BookSearch(ctx context.Context, query string, maxResults int) ([]models.SearchResult, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("book search query is empty")
	}

	u, err := url.Parse(openLibrarySearchURL)
	if err != nil {
		return nil, fmt.Errorf("open library book search: %w", err)
	}
	q := u.Query()
	q.Set("q", query)
	q.Set("limit", strconv.Itoa(openLibraryLimit(maxResults)))
	q.Set("fields", openLibraryFields)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("open library book search: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", openLibraryUA())

	resp, err := e.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("open library book search: %w", err)
	}
	defer closeResponseBody(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("open library book search failed with status %d", resp.StatusCode)
	}
	if isJSONEndpointServingHTML(resp) {
		return nil, fmt.Errorf("open library book search returned HTML")
	}

	var payload olSearchResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, config.MaxBodyBytes)).Decode(&payload); err != nil {
		return nil, fmt.Errorf("open library book search: %w", err)
	}

	results := make([]models.SearchResult, 0, len(payload.Docs))
	for _, doc := range payload.Docs {
		if res, ok := mapOpenLibraryWork(doc); ok {
			results = append(results, res)
		}
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("book search returned no results")
	}
	return results, nil
}

func openLibraryLimit(maxResults int) int {
	if maxResults <= 0 {
		return 5
	}
	if maxResults > openLibraryMaxLimit {
		return openLibraryMaxLimit
	}
	return maxResults
}

func openLibraryUA() string {
	contact := strings.TrimSpace(os.Getenv("OPENLIBRARY_CONTACT"))
	if contact == "" {
		return openLibraryUserAgent
	}
	return "mcp-server-duckduckgo (+https://github.com/maccavelli/mcp-server-duckduckgo; " + contact + ")"
}

func openLibraryWorkURL(key string) (string, bool) {
	if strings.HasPrefix(key, "/works/") && len(key) > len("/works/") {
		return "https://openlibrary.org" + key, true
	}
	if olBareWorkKey.MatchString(key) {
		return "https://openlibrary.org/works/" + key, true
	}
	return "", false
}

func mapOpenLibraryWork(doc olWorkDoc) (models.SearchResult, bool) {
	title := strings.TrimSpace(doc.Title)
	if title == "" {
		return models.SearchResult{}, false
	}
	workURL, ok := openLibraryWorkURL(strings.TrimSpace(doc.Key))
	if !ok {
		return models.SearchResult{}, false
	}
	if subtitle := strings.TrimSpace(doc.Subtitle); subtitle != "" {
		title = title + ": " + subtitle
	}

	authors := make([]string, 0, len(doc.AuthorName))
	for _, name := range doc.AuthorName {
		name = strings.TrimSpace(name)
		if name != "" {
			authors = append(authors, name)
		}
	}
	author := strings.Join(authors, ", ")

	var parts []string
	if author != "" {
		parts = append(parts, author)
	}
	date := ""
	if doc.FirstPublishYear > 0 {
		date = strconv.Itoa(doc.FirstPublishYear)
		parts = append(parts, date)
	}
	if pages := formatPages(doc.PagesMedian); pages != "" {
		parts = append(parts, pages+" pages")
	}
	if doc.EditionCount > 0 {
		parts = append(parts, strconv.Itoa(doc.EditionCount)+" editions")
	}
	line := strings.Join(parts, " · ")

	thumbnail := ""
	if doc.CoverI > 0 {
		thumbnail = "https://covers.openlibrary.org/b/id/" + strconv.Itoa(doc.CoverI) + "-M.jpg"
	}

	return models.SearchResult{
		Title:       truncate(title, config.MaxSnippetLength),
		URL:         truncate(workURL, config.MaxSnippetLength),
		Description: truncate(line, config.MaxSnippetLength),
		Author:      truncate(author, config.MaxSnippetLength),
		Date:        date,
		Thumbnail:   thumbnail,
		Source:      "Open Library",
		Info:        truncate(line, config.MaxSnippetLength),
	}, true
}

func formatPages(n float64) string {
	if n <= 0 {
		return ""
	}
	if n == float64(int64(n)) {
		return strconv.FormatInt(int64(n), 10)
	}
	return strconv.FormatFloat(n, 'f', -1, 64)
}
