package engine

import (
	"io"
	"log/slog"
)

const (
	providerBingImages   = "Bing Images"
	providerGoogleImages = "Google Images"
	htmlTagArticle       = "article"
)

func closeResponseBody(body io.Closer) {
	if body == nil {
		return
	}
	if err := body.Close(); err != nil {
		slog.Debug("failed to close response body", "error", err)
	}
}
