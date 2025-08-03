package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
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

// @Summary Redirect to Auth0 login page
// @Description Redirects the user to Auth0 for authentication
// @Tags auth
// @Produce json
// @Success 307 {string} string "Redirect to Auth0"
// @Router /api/login [get]
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

// @Summary Handle Auth0 callback
// @Description Process the callback from Auth0 after user authentication
// @Tags auth
// @Produce json
// @Param code query string true "Authorization code from Auth0"
// @Success 200 {object} controllers.TokenResponse "Authentication successful"
// @Failure 500 {object} object "Internal server error"
// @Router /api/callback [get]
func Callback(c *gin.Context) {
	code := c.Query("code")
	domain := os.Getenv("AUTH0_DOMAIN")
	clientID := os.Getenv("AUTH0_CLIENT_ID")
	clientSecret := os.Getenv("AUTH0_CLIENT_SECRET")
	redirectURI := os.Getenv("AUTH0_CALLBACK_URL")
	scheme := os.Getenv("AUTH0_SCHEME")
	if scheme == "" {
		scheme = "https"
	}
	// test redeploy comment

	tokenURL := scheme + "://" + domain + "/oauth/token"

	// Validate the URL to prevent potential security issues
	if _, err := url.Parse(tokenURL); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid token URL"})
		return
	}

	reqBody := map[string]string{
		"grant_type":    "authorization_code",
		"client_id":     clientID,
		"client_secret": clientSecret,
		"code":          code,
		"redirect_uri":  redirectURI,
	}
	payload, err := json.Marshal(reqBody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to marshal request body"})
		return
	}
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

	// Set a cookie with the ID token (or access token, as needed)
	// Secure, HttpOnly, and SameSite options are recommended for production
	c.SetCookie(
		"id_token", tr.IDToken,
		tr.ExpiresIn, "/", "nofeed.zone", true, false)

	// Redirect to frontend
	c.Redirect(http.StatusTemporaryRedirect, "https://nofeed.zone")
}

// @Summary Logout user
// @Description Logs out the user by clearing the session cookie and redirecting to Auth0 logout
// @Tags auth
// @Produce json
// @Success 307 {string} string "Redirect to Auth0 logout"
// @Router /api/logout [get]
func Logout(c *gin.Context) {
	domain := os.Getenv("AUTH0_DOMAIN")
	clientID := os.Getenv("AUTH0_CLIENT_ID")
	returnTo := os.Getenv("AUTH0_LOGOUT_RETURN_URL")

	// Clear the authentication cookie
	c.SetCookie(
		"id_token", "",
		-1, "/", "nofeed.zone", true, true) // httpOnly=true

	// Validate required env vars
	if domain == "" || clientID == "" || returnTo == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "logout env vars not set"})
		return
	}

	// Construct the Auth0 logout URL with federated logout to clear all sessions
	logoutURL := fmt.Sprintf(
		"https://%s/v2/logout?client_id=%s&returnTo=%s&federated",
		domain,
		clientID,
		url.QueryEscape(returnTo),
	)
	c.Redirect(http.StatusTemporaryRedirect, logoutURL)
}

// @Summary Check authentication status
// @Description Check if the user is authenticated via cookie
// @Tags auth
// @Produce json
// @Success 200 {object} object "User is authenticated"
// @Failure 401 {object} object "User is not authenticated"
// @Router /api/me [get]
func CheckAuth(c *gin.Context) {
	// Check for id_token cookie
	if cookie, err := c.Cookie("id_token"); err == nil && cookie != "" {
		// Token exists, user is authenticated
		c.JSON(http.StatusOK, gin.H{
			"authenticated": true,
			"message":       "User is authenticated",
		})
		return
	}

	// No valid cookie found
	c.JSON(http.StatusUnauthorized, gin.H{
		"authenticated": false,
		"message":       "User is not authenticated",
	})
}
