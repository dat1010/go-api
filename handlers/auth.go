package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/dat1010/go-api/platform/auth0"
)

// Home shows a login link.
func Home(c *gin.Context) {
	state := uuid.NewString()
	// you’ll want to save “state” in a cookie/session for real-world CSRF protection
	c.HTML(http.StatusOK, "index.html", gin.H{
		"LoginURL": auth0.LoginURL(state),
	})
}

// Login just redirects to Auth0
func Login(c *gin.Context) {
	state := uuid.NewString()
	c.Redirect(http.StatusTemporaryRedirect, auth0.LoginURL(state))
}

// Callback processes Auth0’s response.
func Callback(c *gin.Context) {
	code := c.Query("code")
	token, err := auth0.ExchangeCode(code)
	if err != nil {
		c.String(http.StatusInternalServerError, "token exchange error: %v", err)
		return
	}
	// For now just dump the ID token on screen
	idToken, _ := token.Extra("id_token").(string)
	c.JSON(http.StatusOK, gin.H{
		"access_token": token.AccessToken,
		"id_token":     idToken,
	})
}
