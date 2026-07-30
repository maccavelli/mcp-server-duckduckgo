package store

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/tidwall/buntdb"
)

func setupTestDB(t *testing.T) *buntdb.DB {
	t.Helper()
	db, err := buntdb.Open(":memory:")
	if err != nil {
		t.Fatalf("failed to open memory db: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})
	return db
}

func TestBuntDBJar_SetAndGetCookies(t *testing.T) {
	db := setupTestDB(t)
	jar := NewBuntDBJar(db)

	u, _ := url.Parse("https://example.com/path")

	// Arrange: A basic cookie
	c1 := &http.Cookie{
		Name:     "session_id",
		Value:    "12345",
		Domain:   "example.com",
		Path:     "/",
		Expires:  time.Now().Add(1 * time.Hour),
		Secure:   true,
		HttpOnly: true,
	}

	// Act: Set the cookie
	jar.SetCookies(u, []*http.Cookie{c1})

	// Assert: Retrieve the cookie
	cookies := jar.Cookies(u)
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}
	if cookies[0].Name != "session_id" || cookies[0].Value != "12345" {
		t.Errorf("unexpected cookie content: %+v", cookies[0])
	}
}

func TestBuntDBJar_CookiesFilterBySecureAndPath(t *testing.T) {
	db := setupTestDB(t)
	jar := NewBuntDBJar(db)

	uSet, _ := url.Parse("https://example.com/secured")

	cSecure := &http.Cookie{
		Name:   "secure_cookie",
		Value:  "sec",
		Domain: "example.com",
		Path:   "/",
		Secure: true,
	}
	cPath := &http.Cookie{
		Name:   "path_cookie",
		Value:  "pth",
		Domain: "example.com",
		Path:   "/secured",
	}

	jar.SetCookies(uSet, []*http.Cookie{cSecure, cPath})

	t.Run("https matching path", func(t *testing.T) {
		uMatch, _ := url.Parse("https://example.com/secured/resource")
		cookies := jar.Cookies(uMatch)
		if len(cookies) != 2 {
			t.Errorf("expected 2 cookies, got %d", len(cookies))
		}
	})

	t.Run("http drops secure cookie", func(t *testing.T) {
		uHttp, _ := url.Parse("http://example.com/secured")
		cookies := jar.Cookies(uHttp)
		if len(cookies) != 1 || cookies[0].Name != "path_cookie" {
			t.Errorf("expected only path_cookie, got %v", cookies)
		}
	})

	t.Run("path mismatch drops path cookie", func(t *testing.T) {
		uPath, _ := url.Parse("https://example.com/other")
		cookies := jar.Cookies(uPath)
		if len(cookies) != 1 || cookies[0].Name != "secure_cookie" {
			t.Errorf("expected only secure_cookie, got %v", cookies)
		}
	})
}

func TestBuntDBJar_SetCookies_DeleteExpired(t *testing.T) {
	db := setupTestDB(t)
	jar := NewBuntDBJar(db)

	u, _ := url.Parse("https://example.com")

	// Set initial valid cookie
	c1 := &http.Cookie{
		Name:   "expiring",
		Value:  "val",
		Domain: "example.com",
	}
	jar.SetCookies(u, []*http.Cookie{c1})
	if len(jar.Cookies(u)) != 1 {
		t.Fatal("expected cookie to be set")
	}

	// Set same cookie with negative MaxAge to delete
	c2 := &http.Cookie{
		Name:   "expiring",
		Value:  "val",
		Domain: "example.com",
		MaxAge: -1,
	}
	jar.SetCookies(u, []*http.Cookie{c2})

	if len(jar.Cookies(u)) != 0 {
		t.Error("expected cookie to be deleted")
	}
}

func TestGetCookieDomains(t *testing.T) {
	domains := getCookieDomains("sub.example.com")
	expected := []string{"sub.example.com", ".sub.example.com", "example.com", ".example.com"}

	if len(domains) != len(expected) {
		t.Fatalf("expected %d domains, got %d", len(expected), len(domains))
	}
	for i, d := range domains {
		if d != expected[i] {
			t.Errorf("index %d: expected %s, got %s", i, expected[i], d)
		}
	}
}
