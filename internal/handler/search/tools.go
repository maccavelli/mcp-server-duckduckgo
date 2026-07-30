package search

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/maccavelli/mcplib"

	"github.com/maccavelli/mcp-server-duckduckgo/internal/engine"
	"github.com/maccavelli/mcp-server-duckduckgo/internal/models"
	"github.com/maccavelli/mcp-server-duckduckgo/internal/registry"
	"github.com/maccavelli/mcp-server-duckduckgo/internal/util"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const searchTypeWeb = "web"

// SearchEngine defines the interface for engine searches.
type SearchEngine interface {
	WebSearch(ctx context.Context, query string, maxResults int) ([]models.SearchResult, error)
	NewsSearch(ctx context.Context, query string, maxResults int) ([]models.SearchResult, error)
	BookSearch(ctx context.Context, query string, maxResults int) ([]models.SearchResult, error)
	FetchContent(ctx context.Context, rawURL string) (*engine.FetchedContent, error)
}

// SearchTool implements Tool for various search types.
type SearchTool struct {
	Engine     SearchEngine
	Type       string
	SearchFunc func(context.Context, string, int) ([]models.SearchResult, error)
	Desc       string
}

func (t *SearchTool) Name() string {
	return fmt.Sprintf("search_%s", t.Type)
}

type SearchInput struct {
	util.UniversalBaseInput
	Query        string `json:"query" jsonschema:"The primary search query string to execute"`
	MaxResults   int    `json:"max_results,omitempty" jsonschema:"The maximum number of search results to return. Default is 5."`
	Format       string `json:"format,omitempty" jsonschema:"Output format: 'hybrid', 'json', or 'markdown'.,enum=hybrid,enum=json,enum=markdown"`
	FetchContent bool   `json:"fetch_content,omitempty" jsonschema:"When true, fetches the full page content for each result URL using the hardened HTTP client. Default is false."`
}

func (t *SearchTool) Register(s *mcp.Server) {
	mcplib.HardenedAddTool(s, &mcp.Tool{
		Name:        t.Name(),
		Description: t.Desc,
	}, t.Handle)
}

func (t *SearchTool) Handle(ctx context.Context, request *mcp.CallToolRequest, input SearchInput) (*mcp.CallToolResult, any, error) {
	slog.Info("[BACKPLANE] [SearchTool] executing search action", "type", t.Type, "query", input.Query, "format", input.Format)
	if input.MaxResults <= 0 {
		input.MaxResults = 5
	}
	if input.Format == "" {
		input.Format = os.Getenv("DDG_DEFAULT_FORMAT")
		if input.Format == "" {
			input.Format = "json"
		}
	}

	results, err := t.SearchFunc(ctx, input.Query, input.MaxResults)
	if err != nil {
		res := &mcp.CallToolResult{}
		res.SetError(err)
		return res, nil, nil
	}

	// When fetch_content is enabled, fetch each result's page content concurrently.
	if input.FetchContent && len(results) > 0 {
		results = fetchResultContents(ctx, t.Engine, results)
	}

	response := models.SearchResponse{
		Summary: fmt.Sprintf("Found %d %s results for '%s'", len(results), t.Type, input.Query),
	}
	response.Data.Type = t.Type
	response.Data.Metadata = &models.SearchMetadata{
		Query:      input.Query,
		TotalCount: len(results),
		SearchType: t.Type,
	}
	response.Data.Results = results

	switch input.Format {
	case "markdown":
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: response.ToMarkdown()}},
		}, nil, nil
	case "hybrid":
		response.ResultsMD = response.ToMarkdown()
		return &mcp.CallToolResult{}, response, nil
	default: // json
		return &mcp.CallToolResult{}, response, nil
	}
}

// Register adds the search tools to the registry.
func Register(searchEngine SearchEngine) {
	registry.Global.Register(&SearchTool{
		Engine:     searchEngine,
		Type:       searchTypeWeb,
		SearchFunc: searchEngine.WebSearch,
		Desc:       "[DIRECTIVE: General Internet Discovery] Comprehensive retrieval utilizing the DuckDuckGo index to extract current, real-time web results and engine queries. Keywords: web, internet, websites, online, general-search, urls, text, browse",
	})
	registry.Global.Register(&SearchTool{
		Engine:     searchEngine,
		Type:       "news",
		SearchFunc: searchEngine.NewsSearch,
		Desc:       "[DIRECTIVE: Live Events and Journalism] Preferred choice for breaking developments, real-time updates, and current timeline information. Keywords: news, press, breaking, events, articles, temporal, journalism, headlines",
	})
	registry.Global.Register(&SearchTool{
		Engine:     searchEngine,
		Type:       "books",
		SearchFunc: searchEngine.BookSearch,
		Desc:       "[DIRECTIVE: Academic and Scholarship Lookup] Specialized retrieval of authoritative literature, academic citations, and literary metadata. Keywords: books, academic, scholars, publications, isbn, citations, authors, reading",
	})
	registry.Global.Register(&FetchURLTool{Engine: searchEngine})
}

// fetchResultContents concurrently fetches page content for each search result.
func fetchResultContents(ctx context.Context, eng SearchEngine, results []models.SearchResult) []models.SearchResult {
	var wg sync.WaitGroup
	enriched := make([]models.SearchResult, len(results))
	copy(enriched, results)

	for i := range enriched {
		if enriched[i].URL == "" {
			continue
		}
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			fetched, err := eng.FetchContent(ctx, enriched[idx].URL)
			if err != nil {
				slog.Debug("fetch_content failed for result", "url", enriched[idx].URL, "error", err)
				return
			}
			if fetched.Content != "" {
				enriched[idx].Content = fetched.Content
			}
			if fetched.Title != "" && enriched[idx].Title == "" {
				enriched[idx].Title = fetched.Title
			}
		}(i)
	}
	wg.Wait()
	return enriched
}

// FetchURLTool is a standalone tool for fetching content from any URL.
type FetchURLTool struct {
	Engine SearchEngine
}

type FetchURLInput struct {
	util.UniversalBaseInput
	URL string `json:"url" jsonschema:"The URL to fetch content from"`
}

func (t *FetchURLTool) Name() string { return "fetch_url" }

func (t *FetchURLTool) Register(s *mcp.Server) {
	mcplib.HardenedAddTool(s, &mcp.Tool{
		Name:        t.Name(),
		Description: "[DIRECTIVE: Deep Content Retrieval] Fetches and extracts clean, readable text content from any URL using the hardened HTTP client with cookie persistence, UA rotation, and CAPTCHA detection. Use after search to read full page content. Keywords: fetch, read, extract, scrape, content, page, article, url",
	}, t.Handle)
}

func (t *FetchURLTool) Handle(ctx context.Context, request *mcp.CallToolRequest, input FetchURLInput) (*mcp.CallToolResult, any, error) {
	slog.Info("[BACKPLANE] [FetchURL] fetching content", "url", input.URL)

	fetched, err := t.Engine.FetchContent(ctx, input.URL)
	if err != nil {
		res := &mcp.CallToolResult{}
		res.SetError(err)
		return res, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{
			Text: fmt.Sprintf("# %s\n\nSource: %s\n\n%s", fetched.Title, fetched.URL, fetched.Content),
		}},
	}, nil, nil
}
