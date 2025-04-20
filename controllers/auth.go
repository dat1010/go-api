package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	IDToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

func Login(c *gin.Context) {
	domain := os.Getenv("AUTH0_DOMAIN")
	clientID := os.Getenv("AUTH0_CLIENT_ID")
	redirectURI := os.Getenv("AUTH0_CALLBACK_URL")
	// TODO: generate & verify a real state in production
	state := "example-state"

	authURL := "https://" + domain + "/authorize" +
		"?response_type=code" +
		"&client_id=" + clientID +
		"&redirect_uri=" + redirectURI +
		"&scope=openid%20profile%20email" +
		"&state=" + state

	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

func Callback(c *gin.Context) {
	code := c.Query("code")
	domain := os.Getenv("AUTH0_DOMAIN")
	clientID := os.Getenv("AUTH0_CLIENT_ID")
	clientSecret := os.Getenv("AUTH0_CLIENT_SECRET")
	redirectURI := os.Getenv("AUTH0_CALLBACK_URL")

	tokenURL := "https://" + domain + "/oauth/token"
	reqBody := map[string]string{
		"grant_type":    "authorization_code",
		"client_id":     clientID,
		"client_secret": clientSecret,
		"code":          code,
		"redirect_uri":  redirectURI,
	}
	payload, _ := json.Marshal(reqBody)
	resp, err := http.Post(tokenURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	var tr TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tr)
}
