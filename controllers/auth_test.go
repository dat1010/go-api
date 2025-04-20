package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestLoginRedirect(t *testing.T) {
	os.Setenv("AUTH0_DOMAIN", "dev‑abcd1234.us.auth0.com")
	os.Setenv("AUTH0_CLIENT_ID", "myclientid")
	os.Setenv("AUTH0_CALLBACK_URL", "http://localhost:8080/api/callback")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	Login(c)

	if w.Code != http.StatusTemporaryRedirect {
		t.Fatalf("expected 302, got %d", w.Code)
	}
	loc := w.Header().Get("Location")
	if !strings.HasPrefix(loc, "https://dev‑abcd1234.us.auth0.com/authorize") {
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

	req := httptest.NewRequest("GET", "/callback?code=foo", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	Callback(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var tr TokenResponse
	if err := json.Unmarshal(w.Body.Bytes(), &tr); err != nil {
		t.Fatal(err)
	}
	if tr.AccessToken != "A" || tr.IDToken != "B" {
		t.Errorf("unexpected token payload: %+v", tr)
	}
}
