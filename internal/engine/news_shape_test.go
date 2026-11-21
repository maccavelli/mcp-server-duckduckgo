package engine

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
)

func runNewsShapeTests(t *testing.T) {
	t.Run("query_shape", testNewsQueryShape)
	t.Run("empty_query", testNewsEmptyQuery)
	t.Run("excerpt_plain", testNewsExcerptPlain)
	t.Run("no_rows_skips_bing", testNewsNoRowsSkipsBing)
	t.Run("rows_skip_bing", testNewsRowsSkipBing)
}

type newsQueryCase struct {
	name   string
	query  string
	wantQ  string
	wantDF string
}

func testNewsQueryShape(t *testing.T) {
	cases := []newsQueryCase{
		{"keyword", "Federal Reserve interest rate", "Federal Reserve interest rate", ""},
		{"today", "Federal Reserve interest rate today", "Federal Reserve interest rate", "d"},
		{"right_now", "Federal Reserve interest rate right now", "Federal Reserve interest rate", "d"},
		{"this_week", "Federal Reserve interest rate this week", "Federal Reserve interest rate", "w"},
		{"this_month", "Federal Reserve interest rate this month", "Federal Reserve interest rate", "m"},
		{"question", "what is the federal reserve doing with interest rates", "what is the federal reserve doing with interest rates", ""},
		{"question_now", "what is the federal reserve doing with interest rates right now", "what is the federal reserve doing with interest rates", "d"},
		{"leading_today", "Today in Congress", "in Congress", "d"},
		{"possessive", "rates today's decision", "rates decision", "d"},
		{"spaces", "  Federal   Reserve  ", "Federal Reserve", ""},
		{"both_windows", "rates today this week", "rates", "d"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			checkNewsQueryShape(t, tc)
		})
	}
}

func checkNewsQueryShape(t *testing.T, tc newsQueryCase) {
	t.Helper()
	var mu sync.Mutex
	var vqdQ, newsQ, newsDF, newsNoamp, newsP, newsL, newsO string
	sawNews := false
	client := newMockClient(func(req *http.Request) (*http.Response, error) {
		mu.Lock()
		defer mu.Unlock()
		if strings.Contains(req.URL.Path, "news.js") {
			sawNews = true
			q := req.URL.Query()
			newsQ = q.Get("q")
			if q.Has("df") {
				newsDF = q.Get("df")
			}
			newsNoamp = q.Get("noamp")
			newsP = q.Get("p")
			newsL = q.Get("l")
			newsO = q.Get("o")
			body := `{"results":[{"title":"T","url":"https://example.com/n","excerpt":"e","source":"s"}]}`
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body))}, nil
		}
		if strings.Contains(req.URL.Host, "duckduckgo.com") {
			vqdQ = req.URL.Query().Get("q")
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`vqd='123'`))}, nil
		}
		return &http.Response{StatusCode: 404, Body: io.NopCloser(strings.NewReader(""))}, nil
	})
	e := newTestEngine(t)
	e.Client = client
	if _, err := e.NewsSearch(context.Background(), tc.query, 1); err != nil {
		t.Fatalf("search: %v", err)
	}
	mu.Lock()
	defer mu.Unlock()
	if !sawNews {
		t.Fatal("no news.js request")
	}
	if newsNoamp != "1" || newsP != "-1" || newsL != "us-en" || newsO != "json" {
		t.Fatalf("params noamp=%q p=%q l=%q o=%q", newsNoamp, newsP, newsL, newsO)
	}
	if newsQ != tc.wantQ || vqdQ != tc.wantQ {
		t.Fatalf("q news=%q vqd=%q want %q", newsQ, vqdQ, tc.wantQ)
	}
	if newsDF != tc.wantDF {
		t.Fatalf("df=%q want %q", newsDF, tc.wantDF)
	}
}

func testNewsEmptyQuery(t *testing.T) {
	for _, q := range []string{"today", "today this week", "today's"} {
		t.Run(q, func(t *testing.T) {
			client := newMockClient(func(req *http.Request) (*http.Response, error) {
				t.Errorf("Do called for %q: %s", q, req.URL)
				return &http.Response{StatusCode: 500, Body: io.NopCloser(strings.NewReader(""))}, nil
			})
			e := newTestEngine(t)
			e.Client = client
			_, err := e.NewsSearch(context.Background(), q, 1)
			if err == nil || !strings.Contains(err.Error(), "news query is empty") {
				t.Fatalf("got err %v", err)
			}
		})
	}
}

func testNewsExcerptPlain(t *testing.T) {
	body := `{"results":[{"title":"T","url":"https://example.com/n","excerpt":"Tom&#x27;s <b>Fed</b>","source":"s"}]}`
	client := newMockClient(func(req *http.Request) (*http.Response, error) {
		if strings.Contains(req.URL.Path, "news.js") {
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body))}, nil
		}
		if strings.Contains(req.URL.Host, "duckduckgo.com") {
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`vqd='123'`))}, nil
		}
		return &http.Response{StatusCode: 404, Body: io.NopCloser(strings.NewReader(""))}, nil
	})
	e := newTestEngine(t)
	e.Client = client
	results, err := e.NewsSearch(context.Background(), "Federal Reserve", 1)
	if err != nil || len(results) != 1 {
		t.Fatalf("search: %v len=%d", err, len(results))
	}
	if results[0].Description != "Tom's Fed" {
		t.Fatalf("description %q", results[0].Description)
	}
}

func testNewsNoRowsSkipsBing(t *testing.T) {
	hosts := newsHosts(t, `{"results":[]}`, `<html><body><p>no cards</p></body></html>`)
	for _, h := range hosts {
		if strings.Contains(h, "bing.com") {
			t.Fatalf("bing host %q in %v", h, hosts)
		}
	}
}

func testNewsRowsSkipBing(t *testing.T) {
	body := `{"results":[{"title":"Row","url":"https://example.com/row","excerpt":"e","source":"s"}]}`
	var mu sync.Mutex
	var hosts []string
	count := 0
	client := newMockClient(func(req *http.Request) (*http.Response, error) {
		mu.Lock()
		hosts = append(hosts, req.URL.Host)
		count++
		n := count
		mu.Unlock()
		if strings.Contains(req.URL.Host, "duckduckgo.com") {
			if n == 1 || !strings.Contains(req.URL.Path, "news.js") {
				return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`vqd='123'`))}, nil
			}
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body))}, nil
		}
		return &http.Response{StatusCode: 404, Body: io.NopCloser(strings.NewReader(""))}, nil
	})
	e := newTestEngine(t)
	e.Client = client
	results, err := e.NewsSearch(context.Background(), "Federal Reserve", 1)
	if err != nil || len(results) != 1 || results[0].Title != "Row" {
		t.Fatalf("search: %v results=%v", err, results)
	}
	mu.Lock()
	defer mu.Unlock()
	for _, h := range hosts {
		if strings.Contains(h, "bing.com") {
			t.Fatalf("bing host %q", h)
		}
	}
}

func newsHosts(t *testing.T, newsBody, googleBody string) []string {
	t.Helper()
	var mu sync.Mutex
	var hosts []string
	var errOut error
	client := newMockClient(func(req *http.Request) (*http.Response, error) {
		mu.Lock()
		hosts = append(hosts, req.URL.Host)
		mu.Unlock()
		if strings.Contains(req.URL.Path, "news.js") {
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(newsBody))}, nil
		}
		if strings.Contains(req.URL.Host, "duckduckgo.com") {
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`vqd='123'`))}, nil
		}
		if strings.Contains(req.URL.RawQuery, "tbm=nws") {
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(googleBody))}, nil
		}
		return &http.Response{StatusCode: 404, Body: io.NopCloser(strings.NewReader(""))}, nil
	})
	e := newTestEngine(t)
	e.Client = client
	_, errOut = e.NewsSearch(context.Background(), "Federal Reserve", 1)
	if errOut == nil {
		t.Fatal("expected an error when both news providers are empty")
	}
	mu.Lock()
	defer mu.Unlock()
	out := append([]string(nil), hosts...)
	return out
}
