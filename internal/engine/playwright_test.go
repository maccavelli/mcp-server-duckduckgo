package engine

import (
	"testing"
)

func TestPlaywrightNilChecks(t *testing.T) {
	// These functions should safely handle nil interfaces without panicking
	stopPlaywright(nil)
	closeBrowser(nil)
	closeBrowserContext(nil)
	closePage(nil)
}
