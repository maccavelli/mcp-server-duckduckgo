package engine

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

const bookFixture = `{
  "numFound": 2,
  "num_found": 2,
  "offset": null,
  "docs": [
    {
      "key": "/works/OL1W",
      "title": "Core Python Programming",
      "subtitle": "A Subtitle",
      "author_name": ["Ada Lovelace", "Grace Hopper"],
      "first_publish_year": 1954,
      "number_of_pages_median": 320,
      "cover_i": 99,
      "edition_count": 12
    },
    {
      "key": "OL2W",
      "title": "Bare Key"
    },
    {
      "key": "/works/OL3W",
      "title": "  "
    }
  ]
}`

func TestBookSearchMapsWorks(t *testing.T) {
	t.Setenv("OPENLIBRARY_CONTACT", "")
	var got *http.Request
	client := newMockClient(func(req *http.Request) (*http.Response, error) {
		got = req
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(bookFixture)),
		}, nil
	})
	e := newTestEngine(t)
	e.Client = client

	results, err := e.BookSearch(context.Background(), "  python programming  ", 10)
	if err != nil {
		t.Fatalf("BookSearch: %v", err)
	}
	if got == nil {
		t.Fatal("expected one request")
	}
	if got.URL.Host != "openlibrary.org" || got.URL.Path != "/search.json" {
		t.Fatalf("url = %s", got.URL.String())
	}
	q := got.URL.Query()
	if q.Get("q") != "python programming" {
		t.Errorf("q = %q", q.Get("q"))
	}
	if q.Get("limit") != "10" {
		t.Errorf("limit = %q", q.Get("limit"))
	}
	const fields = "key,title,subtitle,author_name,first_publish_year,number_of_pages_median,cover_i,edition_count"
	if q.Get("fields") != fields {
		t.Errorf("fields = %q", q.Get("fields"))
	}
	if strings.Contains(q.Get("fields"), "isbn") || strings.Contains(q.Get("fields"), "publisher") || strings.Contains(q.Get("fields"), "language") {
		t.Errorf("fields include a dropped name: %q", q.Get("fields"))
	}
	if got.Header.Get("Accept") != "application/json" {
		t.Errorf("Accept = %q", got.Header.Get("Accept"))
	}
	if got.Header.Get("User-Agent") != openLibraryUserAgent {
		t.Errorf("User-Agent = %q", got.Header.Get("User-Agent"))
	}
	if strings.Contains(got.Header.Get("User-Agent"), "Mozilla") {
		t.Errorf("User-Agent looks like a browser: %q", got.Header.Get("User-Agent"))
	}

	if len(results) != 2 {
		t.Fatalf("len = %d, want 2 (empty title dropped)", len(results))
	}
	first := results[0]
	if first.Title != "Core Python Programming: A Subtitle" {
		t.Errorf("title = %q", first.Title)
	}
	if first.URL != "https://openlibrary.org/works/OL1W" {
		t.Errorf("url = %q", first.URL)
	}
	if !strings.HasPrefix(first.URL, "https://openlibrary.org/works/") {
		t.Errorf("url path = %q", first.URL)
	}
	if first.Author != "Ada Lovelace, Grace Hopper" {
		t.Errorf("author = %q", first.Author)
	}
	if first.Source != "Open Library" {
		t.Errorf("source = %q", first.Source)
	}
	if first.Publisher != "" {
		t.Errorf("publisher = %q", first.Publisher)
	}
	if first.Date != "1954" {
		t.Errorf("date = %q", first.Date)
	}
	if first.Thumbnail != "https://covers.openlibrary.org/b/id/99-M.jpg" {
		t.Errorf("thumbnail = %q", first.Thumbnail)
	}
	if first.Description != first.Info {
		t.Errorf("description %q != info %q", first.Description, first.Info)
	}
	if !strings.Contains(first.Description, "1954") || !strings.Contains(first.Description, "320 pages") || !strings.Contains(first.Description, "12 editions") {
		t.Errorf("description = %q", first.Description)
	}

	second := results[1]
	if second.Title != "Bare Key" || second.URL != "https://openlibrary.org/works/OL2W" {
		t.Errorf("second = %#v", second)
	}
	if second.Thumbnail != "" || second.Author != "" {
		t.Errorf("second optional fields = %#v", second)
	}
}

func TestBookSearchLimitClamp(t *testing.T) {
	tests := []struct {
		name string
		max  int
		want string
	}{
		{name: "zero", max: 0, want: "5"},
		{name: "negative", max: -1, want: "5"},
		{name: "over", max: 100, want: "50"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var limit string
			client := newMockClient(func(req *http.Request) (*http.Response, error) {
				limit = req.URL.Query().Get("limit")
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     http.Header{"Content-Type": []string{"application/json"}},
					Body:       io.NopCloser(strings.NewReader(`{"docs":[{"key":"/works/OL1W","title":"T"}]}`)),
				}, nil
			})
			e := newTestEngine(t)
			e.Client = client
			if _, err := e.BookSearch(context.Background(), "q", tt.max); err != nil {
				t.Fatalf("BookSearch: %v", err)
			}
			if limit != tt.want {
				t.Errorf("limit = %q, want %q", limit, tt.want)
			}
		})
	}
}

func TestBookSearchBlankQuerySkipsNetwork(t *testing.T) {
	for _, query := range []string{"", "   "} {
		calls := 0
		client := newMockClient(func(req *http.Request) (*http.Response, error) {
			calls++
			return nil, http.ErrHandlerTimeout
		})
		e := newTestEngine(t)
		e.Client = client
		_, err := e.BookSearch(context.Background(), query, 5)
		if err == nil || !strings.Contains(err.Error(), "book search query is empty") {
			t.Fatalf("query %q: %v", query, err)
		}
		if calls != 0 {
			t.Fatalf("query %q issued %d requests", query, calls)
		}
	}
}

func TestBookSearchContactUserAgent(t *testing.T) {
	t.Setenv("OPENLIBRARY_CONTACT", " dev@example.com ")
	var ua string
	client := newMockClient(func(req *http.Request) (*http.Response, error) {
		ua = req.Header.Get("User-Agent")
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{"docs":[{"key":"/works/OL1W","title":"T"}]}`)),
		}, nil
	})
	e := newTestEngine(t)
	e.Client = client
	if _, err := e.BookSearch(context.Background(), "q", 1); err != nil {
		t.Fatal(err)
	}
	want := "mcp-server-duckduckgo (+https://github.com/maccavelli/mcp-server-duckduckgo; dev@example.com)"
	if ua != want {
		t.Errorf("User-Agent = %q", ua)
	}
}

func TestBookSearchErrors(t *testing.T) {
	cases := []struct {
		name   string
		status int
		ctype  string
		body   string
		want   string
	}{
		{name: "non-200", status: 429, ctype: "application/json", body: `{"docs":[]}`, want: "status 429"},
		{name: "html", status: 200, ctype: "text/html", body: "<html></html>", want: "returned HTML"},
		{name: "truncated", status: 200, ctype: "application/json", body: `{"docs":[`, want: "unexpected EOF"},
		{name: "empty", status: 200, ctype: "application/json", body: `{"offset":null,"docs":[]}`, want: "book search returned no results"},
		{name: "untitled", status: 200, ctype: "application/json", body: `{"docs":[{"key":"/works/OL9W","title":""}]}`, want: "book search returned no results"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client := newMockClient(func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: tc.status,
					Header:     http.Header{"Content-Type": []string{tc.ctype}},
					Body:       io.NopCloser(strings.NewReader(tc.body)),
				}, nil
			})
			e := newTestEngine(t)
			e.Client = client
			_, err := e.BookSearch(context.Background(), "q", 5)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("got %v, want substring %q", err, tc.want)
			}
		})
	}
}
