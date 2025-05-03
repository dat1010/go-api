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
		-1, "/", "nofeed.zone", true, false)

	// Construct the Auth0 logout URL
	logoutURL := "https://" + domain + "/v2/logout" +
		"?client_id=" + clientID +
		"&returnTo=" + returnTo

	c.Redirect(http.StatusTemporaryRedirect, logoutURL)
}
