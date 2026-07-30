package models

import (
	"testing"
)

func TestToMarkdown(t *testing.T) {
	sr := SearchResponse{}
	sr.Data.Type = "Web"
	sr.Data.Metadata = &SearchMetadata{Query: "test"}
	sr.Data.Results = []SearchResult{{Title: "Test", URL: "http://test", Description: "desc"}}

	md := sr.ToMarkdown()
	if len(md) == 0 {
		t.Error("expected non-empty markdown")
	}

	sr20 := SearchResponse20{}
	sr20.Data.Metadata = &SearchMetadata{SearchType: "Image", Query: "test"}
	sr20.Data.Results = []MediaResult20{{Title: "Img", MediaURL: "http://img", PageURL: "http://page"}}

	md2 := sr20.ToMarkdown()
	if len(md2) == 0 {
		t.Error("expected non-empty markdown")
	}
}
