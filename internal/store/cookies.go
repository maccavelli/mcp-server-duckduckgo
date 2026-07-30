package store

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/tidwall/buntdb"
)

const (
	// cookiePrefix is the BuntDB key prefix for stored cookies.
	cookiePrefix = "cookie:"

	// cookieMaxTTL caps the TTL for any persisted cookie.
	cookieMaxTTL = 24 * time.Hour
)

// StoredCookie contains only the exported fields of *http.Cookie that are
// safe for JSON serialization. The unexported fields (Raw, RawExpires,
// Unparsed) are parser artifacts and are NOT used by http.Client.Do when
// attaching cookies to outgoing requests via req.AddCookie().
type StoredCookie struct {
	Name     string        `json:"name"`
	Value    string        `json:"value"`
	Domain   string        `json:"domain"`
	Path     string        `json:"path"`
	Expires  time.Time     `json:"expires"`
	MaxAge   int           `json:"max_age"`
	Secure   bool          `json:"secure"`
	HttpOnly bool          `json:"http_only"`
	SameSite http.SameSite `json:"same_site"`
}

// toHTTPCookie reconstructs an *http.Cookie from the stored exported fields.
func (sc *StoredCookie) toHTTPCookie() *http.Cookie {
	sameSite := sc.SameSite
	if sameSite == 0 {
		sameSite = http.SameSiteLaxMode
	}
	//nolint:gosec // G124: faithfully restoring cookie attributes from persistent jar storage
	return &http.Cookie{
		Name:     sc.Name,
		Value:    sc.Value,
		Domain:   sc.Domain,
		Path:     sc.Path,
		Expires:  sc.Expires,
		MaxAge:   sc.MaxAge,
		Secure:   sc.Secure,
		HttpOnly: sc.HttpOnly,
		SameSite: sameSite,
	}
}

// fromHTTPCookie extracts the serializable fields from an *http.Cookie.
func fromHTTPCookie(c *http.Cookie) StoredCookie {
	return StoredCookie{
		Name:     c.Name,
		Value:    c.Value,
		Domain:   c.Domain,
		Path:     c.Path,
		Expires:  c.Expires,
		MaxAge:   c.MaxAge,
		Secure:   c.Secure,
		HttpOnly: c.HttpOnly,
		SameSite: c.SameSite,
	}
}

// cookieKey builds a BuntDB key for a cookie: "cookie:{domain}:{name}".
func cookieKey(domain, name string) string {
	return cookiePrefix + domain + ":" + name
}

// BuntDBJar implements http.CookieJar backed by BuntDB. The http.Client
// automatically calls SetCookies on responses and Cookies on requests,
// so no manual Set-Cookie header parsing is required.
type BuntDBJar struct {
	db *buntdb.DB
}

// NewBuntDBJar creates a new CookieJar backed by the given BuntDB instance.
func NewBuntDBJar(db *buntdb.DB) *BuntDBJar {
	return &BuntDBJar{db: db}
}

// SetCookies persists cookies received from the given URL into BuntDB.
// Called automatically by http.Client after each response.
// BuntDB's MVCC guarantees atomic writes — no partial cookie state.
func (j *BuntDBJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	if len(cookies) == 0 {
		return
	}

	domain := u.Hostname()

	if err := j.db.Update(func(tx *buntdb.Tx) error {
		for _, c := range cookies {
			sc := fromHTTPCookie(c)
			if sc.Domain == "" {
				sc.Domain = domain
			}

			if c.MaxAge < 0 || (!c.Expires.IsZero() && time.Until(c.Expires) <= 0) {
				if _, err := tx.Delete(cookieKey(sc.Domain, sc.Name)); err != nil {
					slog.Debug("failed to delete expired cookie", "domain", sc.Domain, "name", sc.Name, "error", err)
				}
				continue
			}

			data, err := json.Marshal(sc)
			if err != nil {
				continue
			}

			ttl := cookieMaxTTL
			if !c.Expires.IsZero() {
				remaining := time.Until(c.Expires)
				if remaining > 0 && remaining < ttl {
					ttl = remaining
				}
			}

			if _, _, err := tx.Set(cookieKey(sc.Domain, sc.Name), string(data), &buntdb.SetOptions{
				Expires: true,
				TTL:     ttl,
			}); err != nil {
				slog.Debug("failed to persist cookie", "domain", sc.Domain, "name", sc.Name, "error", err)
			}
		}
		return nil
	}); err != nil {
		slog.Warn("failed to update cookie jar", "error", err)
	}
}

// getCookieDomains generates all valid parent domains for cookie matching.
func getCookieDomains(hostname string) []string {
	domains := []string{hostname}
	if !strings.HasPrefix(hostname, ".") {
		domains = append(domains, "."+hostname)
	}
	parts := strings.Split(hostname, ".")
	for i := 1; i < len(parts)-1; i++ {
		parent := strings.Join(parts[i:], ".")
		domains = append(domains, parent, "."+parent)
	}
	return domains
}

// Cookies retrieves all stored cookies matching the given URL's domain.
// Called automatically by http.Client before each request.
// Uses BuntDB's AscendKeys for efficient prefix scanning.
func (j *BuntDBJar) Cookies(u *url.URL) []*http.Cookie {
	var cookies []*http.Cookie
	domains := getCookieDomains(u.Hostname())

	if err := j.db.View(func(tx *buntdb.Tx) error {
		for _, domain := range domains {
			prefix := cookiePrefix + domain + ":"
			if err := tx.AscendKeys(prefix+"*", func(key, val string) bool {
				var sc StoredCookie
				if err := json.Unmarshal([]byte(val), &sc); err != nil {
					return true // skip malformed entries
				}

				c := sc.toHTTPCookie()

				// Filter: skip Secure cookies on non-HTTPS URLs.
				if c.Secure && u.Scheme != "https" {
					return true
				}

				// Filter: skip cookies whose path doesn't match.
				if c.Path != "" && !strings.HasPrefix(u.Path, c.Path) {
					return true
				}

				cookies = append(cookies, c)
				return true
			}); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		slog.Warn("failed to read cookies from jar", "error", err)
	}

	return cookies
}
