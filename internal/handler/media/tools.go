package media

import (
	"context"
	"fmt"
	"os"

	"github.com/maccavelli/mcplib"

	"github.com/maccavelli/mcp-server-duckduckgo/internal/models"
	"github.com/maccavelli/mcp-server-duckduckgo/internal/registry"
	"github.com/maccavelli/mcp-server-duckduckgo/internal/util"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	mediaTypeImages = "images"
	mediaTypeVideos = "videos"
)

// SearchEngine defines the interface for engine searches.
type SearchEngine interface {
	ImageSearch(ctx context.Context, query string, maxResults int) ([]models.SearchResult, error)
	VideoSearch(ctx context.Context, query string, maxResults int) ([]models.SearchResult, error)
}

// MediaTool implements Tool for various media search types.
type MediaTool struct {
	Engine     SearchEngine
	Type       string
	SearchFunc func(context.Context, string, int) ([]models.SearchResult, error)
	Desc       string
}

func (t *MediaTool) Name() string {
	return fmt.Sprintf("search_%s", t.Type)
}

type MediaInput struct {
	util.UniversalBaseInput
	Query      string `json:"query" jsonschema:"The search keywords"`
	MaxResults int    `json:"max_results,omitempty" jsonschema:"Maximum results to return (default 5). Low counts are faster and more token-efficient."`
	Format     string `json:"format,omitempty" jsonschema:"Output format: 'hybrid' (JSON metadata + markdown content), 'json' (pure structured data), or 'markdown' (all markdown formatting).,enum=hybrid,enum=json,enum=markdown"`
}

func (t *MediaTool) Register(s *mcp.Server) {
	mcplib.HardenedAddTool(s, &mcp.Tool{
		Name:        t.Name(),
		Description: t.Desc,
	}, t.Handle)
}

func (t *MediaTool) Handle(ctx context.Context, request *mcp.CallToolRequest, input MediaInput) (*mcp.CallToolResult, any, error) {
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

	mediaResults := make([]models.MediaResult20, 0, len(results))
	for _, r := range results {
		mediaURL := r.ImageURL
		if mediaURL == "" && t.Type == mediaTypeVideos {
			mediaURL = r.URL
		}
		mediaResults = append(mediaResults, models.MediaResult20{
			Title:        r.Title,
			PageURL:      r.URL,
			MediaURL:     mediaURL,
			ThumbnailURL: r.Thumbnail,
			Duration:     r.Duration,
			Publisher:    r.Publisher,
			Source:       r.Source,
		})
	}

	response := models.SearchResponse20{
		Summary: fmt.Sprintf("Found %d %s results for '%s'", len(results), t.Type, input.Query),
	}
	response.Data.Version = "2.0"
	response.Data.Metadata = &models.SearchMetadata{
		Query:      input.Query,
		TotalCount: len(results),
		SearchType: t.Type,
	}
	response.Data.Results = mediaResults

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

// Register adds the media tools to the registry.
func Register(engine SearchEngine) {
	registry.Global.Register(&MediaTool{
		Engine:     engine,
		Type:       mediaTypeImages,
		SearchFunc: engine.ImageSearch,
		Desc:       "[DIRECTIVE: Visual Asset Retrieval] High-priority visual discovery to locate UI/UX inspiration, technical diagrams, photographs, and logos. Keywords: images, pictures, graphics, photos, diagrams, visual, assets, layouts",
	})
	registry.Global.Register(&MediaTool{
		Engine:     engine,
		Type:       mediaTypeVideos,
		SearchFunc: engine.VideoSearch,
		Desc:       "[DIRECTIVE: Multimedia and Demonstration Retrieval] Targeted search for tutorials, walkthroughs, multimedia reports, and movies. Keywords: videos, multimedia, tutorials, walkthroughs, movies, clips, streaming, media",
	})
}
