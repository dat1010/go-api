package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/dat1010/go-api/utils"
	"github.com/gin-gonic/gin"
)

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	IDToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

const (
	accessTokenCookie  = "access_token"
	refreshTokenCookie = "refresh_token"
)

func cookieDomain(c *gin.Context) string {
	domain := strings.TrimSpace(os.Getenv("AUTH0_COOKIE_DOMAIN"))
	if domain != "" {
		return domain
	}
	return ""
}

func cookieSecure() bool {
	secure := strings.TrimSpace(os.Getenv("AUTH0_COOKIE_SECURE"))
	if secure == "" {
		return true
	}
	return strings.EqualFold(secure, "true") || secure == "1"
}

func cookieSameSite(c *gin.Context) {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("AUTH0_COOKIE_SAMESITE"))) {
	case "strict":
		c.SetSameSite(http.SameSiteStrictMode)
	case "none":
		c.SetSameSite(http.SameSiteNoneMode)
	default:
		c.SetSameSite(http.SameSiteLaxMode)
	}
}

func setTokenCookie(c *gin.Context, name, value string, maxAge int) {
	cookieSameSite(c)
	c.SetCookie(name, value, maxAge, "/", cookieDomain(c), cookieSecure(), true)
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
	audience := os.Getenv("AUTH0_AUDIENCE")
	// TODO: generate & verify a real state in production
	state := "example-state"

	authURL := "https://" + domain + "/authorize" +
		"?response_type=code" +
		"&client_id=" + clientID +
		"&redirect_uri=" + redirectURI +
		"&scope=openid%20profile%20email%20offline_access" +
		"&state=" + state

	if audience != "" {
		authURL += "&audience=" + audience
	}

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
	if clientID == "" || clientSecret == "" || domain == "" || redirectURI == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "auth env vars not set"})
		return
	}

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

	if tr.AccessToken == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "missing access token"})
		return
	}

	// Set cookies for access and refresh tokens (if provided)
	setTokenCookie(c, accessTokenCookie, tr.AccessToken, tr.ExpiresIn)
	if tr.RefreshToken != "" {
		// Refresh tokens should generally be long-lived; let Auth0 control expiry server-side.
		setTokenCookie(c, refreshTokenCookie, tr.RefreshToken, 60*60*24*30)
	}

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
	setTokenCookie(c, accessTokenCookie, "", -1)
	setTokenCookie(c, refreshTokenCookie, "", -1)

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
	auth0UserID, ok := utils.GetAuth0UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"authenticated": false,
			"message":       "User is not authenticated",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"authenticated": true,
		"user_id":       auth0UserID,
		"message":       "User is authenticated",
	})
}

// @Summary Refresh access token
// @Description Exchange refresh token for a new access token
// @Tags auth
// @Produce json
// @Success 200 {object} object "Token refreshed"
// @Failure 401 {object} object "Missing refresh token"
// @Failure 500 {object} object "Internal server error"
// @Router /api/refresh [post]
func Refresh(c *gin.Context) {
	refreshToken, err := c.Cookie(refreshTokenCookie)
	if err != nil || refreshToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing refresh token"})
		return
	}

	domain := os.Getenv("AUTH0_DOMAIN")
	clientID := os.Getenv("AUTH0_CLIENT_ID")
	clientSecret := os.Getenv("AUTH0_CLIENT_SECRET")
	scheme := os.Getenv("AUTH0_SCHEME")
	if scheme == "" {
		scheme = "https"
	}
	if clientID == "" || clientSecret == "" || domain == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "auth env vars not set"})
		return
	}

	tokenURL := scheme + "://" + domain + "/oauth/token"
	if _, err := url.Parse(tokenURL); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid token URL"})
		return
	}

	reqBody := map[string]string{
		"grant_type":    "refresh_token",
		"client_id":     clientID,
		"client_secret": clientSecret,
		"refresh_token": refreshToken,
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

	if tr.AccessToken == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "missing access token"})
		return
	}

	setTokenCookie(c, accessTokenCookie, tr.AccessToken, tr.ExpiresIn)
	if tr.RefreshToken != "" {
		setTokenCookie(c, refreshTokenCookie, tr.RefreshToken, 60*60*24*30)
	}

	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, gin.H{"refreshed": true})
}
