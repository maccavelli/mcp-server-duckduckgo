package engine

import (
	"net/http"
)

// mockTransport implements http.RoundTripper for mocking.
type mockTransport struct {
	roundTrip func(*http.Request) (*http.Response, error)
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.roundTrip(req)
}

func newMockClient(roundTrip func(*http.Request) (*http.Response, error)) *http.Client {
	return &http.Client{Transport: &mockTransport{roundTrip}}
}
