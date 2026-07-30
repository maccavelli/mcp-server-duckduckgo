package engine

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/go-rod/stealth"
	"github.com/playwright-community/playwright-go"
)

var pwInstallOnce sync.Once

// ensurePlaywright ensures the driver is installed exactly once globally.
func ensurePlaywright() {
	pwInstallOnce.Do(func() {
		if err := playwright.Install(); err != nil {
			slog.Warn("playwright install failed", "error", err)
		}
	})
}

func stopPlaywright(pw *playwright.Playwright) {
	if pw == nil {
		return
	}
	if err := pw.Stop(); err != nil {
		slog.Debug("failed to stop playwright", "error", err)
	}
}

func closeBrowser(browser playwright.Browser) {
	if browser == nil {
		return
	}
	if err := browser.Close(); err != nil {
		slog.Debug("failed to close browser", "error", err)
	}
}

func closeBrowserContext(bCtx playwright.BrowserContext) {
	if bCtx == nil {
		return
	}
	if err := bCtx.Close(); err != nil {
		slog.Debug("failed to close browser context", "error", err)
	}
}

func closePage(page playwright.Page) {
	if page == nil {
		return
	}
	if err := page.Close(); err != nil {
		slog.Debug("failed to close page", "error", err)
	}
}

// fetchWithPlaywright serves as the Workaround 2 fallback for handling JS-based CAPTCHAs.
// It executes a full headless browser session, injects the go-rod stealth JS payload to
// bypass Turnstile/Datadome protections, and explicitly syncs any acquired clearance cookies
// to the BuntDB store (Workaround 4) for future use by the native HTTP client.
func (e *SearchEngine) fetchWithPlaywright(ctx context.Context, targetURL string) ([]byte, error) {
	ensurePlaywright()

	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("could not start playwright: %w", err)
	}
	defer stopPlaywright(pw)

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: new(true),
	})
	if err != nil {
		return nil, fmt.Errorf("could not launch chromium: %w", err)
	}
	defer closeBrowser(browser)

	// Workaround: aggressively enforce context cancellation
	doneCh := make(chan struct{})
	defer close(doneCh)
	go func() {
		select {
		case <-ctx.Done():
			closeBrowser(browser)
			stopPlaywright(pw)
		case <-doneCh:
		}
	}()

	bCtx, err := browser.NewContext(playwright.BrowserNewContextOptions{
		UserAgent: new(e.store.GetUA(targetURL)),
		Viewport: &playwright.Size{
			Width:  1920,
			Height: 1080,
		},
		Locale:      new("en-US"),
		TimezoneId:  new("America/New_York"),
		ColorScheme: playwright.ColorSchemeDark,
	})
	if err != nil {
		return nil, err
	}
	defer closeBrowserContext(bCtx)

	// Workaround 2: Inject Puppeteer-extra-plugin-stealth equivalents
	// We use the stealth.JS payload from go-rod, which contains all necessary
	// evasions (navigator.webdriver, chrome runtime, webgl masking).
	err = bCtx.AddInitScript(playwright.Script{
		Content: playwright.String(stealth.JS),
	})
	if err != nil {
		return nil, err
	}

	page, err := bCtx.NewPage()
	if err != nil {
		return nil, err
	}
	defer closePage(page)

	// Map context deadline to playwright timeout if possible, else 30s
	timeout := 30000.0
	if deadline, ok := ctx.Deadline(); ok {
		remaining := time.Until(deadline).Milliseconds()
		if remaining > 0 {
			timeout = float64(remaining)
		}
	}
	page.SetDefaultTimeout(timeout)

	resp, err := page.Goto(targetURL, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, fmt.Errorf("playwright got empty response for %s", targetURL)
	}

	// Wait for any dynamically injected Turnstile/Cloudflare challenges to resolve
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(5 * time.Second):
	}

	content, err := page.Content()
	if err != nil {
		return nil, err
	}

	// Workaround 4: Extract clearance cookies and persist them to the BuntDBJar.
	// This ensures that subsequent requests made by the lightweight `http.Client`
	// can bypass the challenge without spinning up the browser again.
	pwCookies, err := bCtx.Cookies()
	if err == nil && e.Client.Jar != nil {
		parsedURL, parseErr := url.Parse(targetURL)
		if parseErr == nil && parsedURL != nil {
			httpCookies := make([]*http.Cookie, 0, len(pwCookies))
			for _, c := range pwCookies {
				//nolint:gosec // G124: syncing browser-acquired cookies with original security attributes
				httpCookies = append(httpCookies, &http.Cookie{
					Name:     c.Name,
					Value:    c.Value,
					Domain:   c.Domain,
					Path:     c.Path,
					Secure:   c.Secure,
					HttpOnly: c.HttpOnly,
					SameSite: http.SameSiteLaxMode,
				})
			}
			e.Client.Jar.SetCookies(parsedURL, httpCookies)
		}
	}

	return []byte(content), nil
}
