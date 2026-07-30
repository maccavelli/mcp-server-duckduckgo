package engine

import (
	"errors"
	"net/http"
	"strings"
)

// ErrCAPTCHADetected signals that a provider returned a challenge page
// instead of search results.
var ErrCAPTCHADetected = errors.New("CAPTCHA or challenge page detected")

// ChallengeType classifies the detected challenge mechanism.
type ChallengeType string

const (
	ChallengeContentType ChallengeType = "content_type_mismatch"
	ChallengeDDGModal    ChallengeType = "ddg_anomaly_modal"
	ChallengeGoogleDiv   ChallengeType = "google_challenge_div"
	ChallengeBingCaptcha ChallengeType = "bing_captcha"
	ChallengeCloudflare  ChallengeType = "cloudflare_turnstile"
)

// isJSONEndpointServingHTML checks the Content-Type of a response that
// SHOULD return JSON (e.g. /i.js, /v.js, /news.js). If the server
// returns text/html instead, this indicates a challenge page was served.
// This is the broadest defensive layer — it catches ALL challenge variants
// regardless of specific HTML structure.
func isJSONEndpointServingHTML(resp *http.Response) bool {
	ct := resp.Header.Get("Content-Type")
	return strings.Contains(ct, "text/html")
}

// detectDDGChallenge checks for DDG-specific challenge indicators in HTML.
func detectDDGChallenge(body []byte) ChallengeType {
	s := string(body)
	if strings.Contains(s, "anomaly-modal") || strings.Contains(s, "Please try again") {
		return ChallengeDDGModal
	}
	if strings.Contains(s, "cf-turnstile") || strings.Contains(s, "challenges.cloudflare.com") {
		return ChallengeCloudflare
	}
	return ""
}

// detectGoogleChallenge checks for Google's JS-based challenge page.
func detectGoogleChallenge(body []byte) ChallengeType {
	s := string(body)
	if strings.Contains(s, "id=\"captcha-form\"") || strings.Contains(s, "g-recaptcha") ||
		strings.Contains(s, "Our systems have detected unusual traffic") {
		return ChallengeGoogleDiv
	}
	return ""
}

// detectBingChallenge checks for Bing's CAPTCHA indicators.
func detectBingChallenge(body []byte) ChallengeType {
	s := string(body)
	if strings.Contains(s, "class=\"captcha\"") || strings.Contains(s, "CaptchaChar") ||
		strings.Contains(s, "bing.com/ck/a?!&&p=captcha") {
		return ChallengeBingCaptcha
	}
	return ""
}
