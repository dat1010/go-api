package controllers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestLoginRedirect(t *testing.T) {
	os.Setenv("AUTH0_DOMAIN", "dev-abcd1234.us.auth0.com")
	os.Setenv("AUTH0_CLIENT_ID", "myclientid")
	os.Setenv("AUTH0_CALLBACK_URL", "http://localhost:8080/api/callback")

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
}

func TestCallbackReturnsToken(t *testing.T) {
	// mock Auth0 token endpoint
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"access_token":"A","id_token":"B","expires_in":3600,"token_type":"Bearer"}`))
	}))
	defer ts.Close()

	// point domain at our test server (strip scheme)
	host := strings.TrimPrefix(ts.URL, "http://")
	os.Setenv("AUTH0_DOMAIN", host)
	os.Setenv("AUTH0_CLIENT_ID", "id")
	os.Setenv("AUTH0_CLIENT_SECRET", "secret")
	os.Setenv("AUTH0_CALLBACK_URL", "http://localhost:8080/api/callback")
	os.Setenv("AUTH0_SCHEME", "http")

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
