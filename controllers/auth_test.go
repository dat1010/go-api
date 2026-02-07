package controllers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/gin-gonic/gin"
)

func TestLoginRedirect(t *testing.T) {
	os.Setenv("AUTH0_DOMAIN", "dev-abcd1234.us.auth0.com")
	os.Setenv("AUTH0_CLIENT_ID", "myclientid")
	os.Setenv("AUTH0_CALLBACK_URL", "http://localhost:8080/api/callback")
	os.Setenv("AUTH0_AUDIENCE", "https://api.example.com")

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/login", nil)
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	Login(c)

	if w.Code != http.StatusTemporaryRedirect {
		t.Fatalf("expected 302, got %d", w.Code)
	}
	loc := w.Header().Get("Location")
	if !strings.HasPrefix(loc, "https://dev-abcd1234.us.auth0.com/authorize") {
		t.Errorf("unexpected redirect URL: %s", loc)
	}
	if !strings.Contains(loc, "audience=https://api.example.com") {
		t.Errorf("audience missing in redirect URL: %s", loc)
	}
}

func TestCallbackReturnsToken(t *testing.T) {
	// mock Auth0 token endpoint
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"access_token":"A","id_token":"B","refresh_token":"R","expires_in":3600,"token_type":"Bearer"}`))
	}))
	defer ts.Close()

	// point domain at our test server (strip scheme)
	host := strings.TrimPrefix(ts.URL, "http://")
	os.Setenv("AUTH0_DOMAIN", host)
	os.Setenv("AUTH0_CLIENT_ID", "id")
	os.Setenv("AUTH0_CLIENT_SECRET", "secret")
	os.Setenv("AUTH0_CALLBACK_URL", "http://localhost:8080/api/callback")
	os.Setenv("AUTH0_SCHEME", "http")
	os.Setenv("AUTH0_COOKIE_DOMAIN", "")

	req := httptest.NewRequest("GET", "/callback?code=foo", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	Callback(c)

	if w.Code != http.StatusTemporaryRedirect {
		t.Fatalf("expected 307, got %d", w.Code)
	}
	loc := w.Header().Get("Location")
	if loc != "https://nofeed.zone" {
		t.Errorf("unexpected redirect location: %s", loc)
	}
}

func TestRefreshReturnsToken(t *testing.T) {
	// mock Auth0 token endpoint
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"access_token":"NEW","refresh_token":"NEWREF","expires_in":3600,"token_type":"Bearer"}`))
	}))
	defer ts.Close()

	// point domain at our test server (strip scheme)
	host := strings.TrimPrefix(ts.URL, "http://")
	os.Setenv("AUTH0_DOMAIN", host)
	os.Setenv("AUTH0_CLIENT_ID", "id")
	os.Setenv("AUTH0_CLIENT_SECRET", "secret")
	os.Setenv("AUTH0_SCHEME", "http")
	os.Setenv("AUTH0_COOKIE_DOMAIN", "")

	req := httptest.NewRequest("POST", "/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "R"})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	Refresh(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	setCookies := w.Header().Values("Set-Cookie")
	if len(setCookies) == 0 {
		t.Fatalf("expected Set-Cookie headers, got none")
	}
	foundAccess := false
	foundRefresh := false
	for _, sc := range setCookies {
		if strings.HasPrefix(sc, "access_token=") {
			foundAccess = true
		}
		if strings.HasPrefix(sc, "refresh_token=") {
			foundRefresh = true
		}
	}
	if !foundAccess || !foundRefresh {
		t.Fatalf("expected access_token and refresh_token cookies, got: %v", setCookies)
	}
}

func TestCheckAuthReturnsUserID(t *testing.T) {
	req := httptest.NewRequest("GET", "/me", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	c.Set("user", validator.RegisteredClaims{Subject: "auth0|testuser"})

	CheckAuth(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), `"authenticated":true`) {
		t.Fatalf("expected authenticated true, got: %s", w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"user_id":"auth0|testuser"`) {
		t.Fatalf("expected user_id, got: %s", w.Body.String())
	}
}
