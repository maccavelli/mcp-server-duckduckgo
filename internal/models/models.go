package models

import (
	"fmt"
	"strings"
)

// SearchResult represents a single search result from DuckDuckGo or mirrors.
type SearchResult struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
	Content     string `json:"content,omitempty"`
	Source      string `json:"source,omitempty"`
	Date        string `json:"date,omitempty"`
	ImageURL    string `json:"image_url,omitempty"`
	Thumbnail   string `json:"thumbnail,omitempty"`
	Duration    string `json:"duration,omitempty"`
	Publisher   string `json:"publisher,omitempty"`
	Author      string `json:"author,omitempty"`
	Info        string `json:"info,omitempty"`
}

// SearchMetadata represents machine-readable metadata about the search.
type SearchMetadata struct {
	Query      string `json:"query"`
	TotalCount int    `json:"total_count"`
	SearchType string `json:"search_type"`
}

// SearchResponse represents the formatted response for the MCP tool.
type SearchResponse struct {
	Summary   string `json:"summary"`
	ResultsMD string `json:"results_md,omitempty"`
	Data      struct {
		Type     string          `json:"type"`
		Metadata *SearchMetadata `json:"metadata"`
		Results  []SearchResult  `json:"results"`
	} `json:"data"`
}

func (r SearchResponse) ToMarkdown() string {
	var res strings.Builder
	fmt.Fprintf(&res, "# %s Search Results for '%s'\n\n", r.Data.Type, r.Data.Metadata.Query)
	for i, item := range r.Data.Results {
		fmt.Fprintf(&res, "### %d. [%s](%s)\n", i+1, item.Title, item.URL)
		if item.Description != "" {
			fmt.Fprintf(&res, "> %s\n\n", item.Description)
		} else {
			res.WriteString("\n")
		}
		if item.Content != "" {
			fmt.Fprintf(&res, "<details>\n<summary>Page Content</summary>\n\n%s\n\n</details>\n\n", item.Content)
		}
	}
	return res.String()
}

// SearchResponse20 represents the Structured JSON 2.0 format for media results.
type SearchResponse20 struct {
	Summary   string `json:"summary"`
	ResultsMD string `json:"results_md,omitempty"`
	Data      struct {
		Version  string          `json:"version"`
		Metadata *SearchMetadata `json:"metadata"`
		Results  []MediaResult20 `json:"results"`
	} `json:"data"`
}

func (r SearchResponse20) ToMarkdown() string {
	var res strings.Builder
	fmt.Fprintf(&res, "# %s Media Results for '%s'\n\n", r.Data.Metadata.SearchType, r.Data.Metadata.Query)
	for i, item := range r.Data.Results {
		fmt.Fprintf(&res, "### %d. %s\n", i+1, item.Title)
		if item.MediaURL != "" {
			fmt.Fprintf(&res, "![%s](%s)\n", item.Title, item.MediaURL)
		}
		fmt.Fprintf(&res, "[View Page](%s)\n\n", item.PageURL)
	}
	return res.String()
}

// MediaResult20 represents a single result in the Structured JSON 2.0 format.
type MediaResult20 struct {
	Title        string `json:"title"`
	PageURL      string `json:"page_url"`
	MediaURL     string `json:"media_url,omitempty"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
	Duration     string `json:"duration,omitempty"`
	Publisher    string `json:"publisher,omitempty"`
	Source       string `json:"source,omitempty"`
}
