package controllers

import (
 	"net/http"
 	"net/http/httptest"
 	"testing"

 	"github.com/gin-gonic/gin"
)

func TestHelloWorld(t *testing.T) {
 	gin.SetMode(gin.TestMode)
 	w := httptest.NewRecorder()
 	c, _ := gin.CreateTestContext(w)
 	req := httptest.NewRequest("GET", "/", nil)
 	c.Request = req
 	HelloWorld(c)
 	if w.Code != http.StatusOK {
 		t.Fatalf("expected status 200, got %d", w.Code)
 	}
 	if body := w.Body.String(); body != "hello" {
 		t.Fatalf("expected body 'hello', got %q", body)
 	}
}
