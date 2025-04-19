package auth0

import (
	"context"
	"errors"
	"os"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

var cfg = &oauth2.Config{
	ClientID:     os.Getenv("AUTH0_CLIENT_ID"),
	ClientSecret: os.Getenv("AUTH0_CLIENT_SECRET"),
	RedirectURL:  os.Getenv("AUTH0_CALLBACK_URL"),
	Scopes:       []string{"openid", "profile", "email"},
	Endpoint: oauth2.Endpoint{
		AuthURL:  "https://" + os.Getenv("AUTH0_DOMAIN") + "/authorize",
		TokenURL: "https://" + os.Getenv("AUTH0_DOMAIN") + "/oauth/token",
	},
}

// Authenticator is used to authenticate our users.
type Authenticator struct {
	*oidc.Provider
	oauth2.Config
}

// New instantiates the *Authenticator.
func New() (*Authenticator, error) {
	provider, err := oidc.NewProvider(
		context.Background(),
		"https://"+os.Getenv("AUTH0_DOMAIN")+"/",
	)
	if err != nil {
		return nil, err
	}

	conf := oauth2.Config{
		ClientID:     os.Getenv("AUTH0_CLIENT_ID"),
		ClientSecret: os.Getenv("AUTH0_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("AUTH0_CALLBACK_URL"),
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile"},
	}

	return &Authenticator{
		Provider: provider,
		Config:   conf,
	}, nil
}

// VerifyIDToken verifies that an *oauth2.Token is a valid *oidc.IDToken.
func (a *Authenticator) VerifyIDToken(ctx context.Context, token *oauth2.Token) (*oidc.IDToken, error) {
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, errors.New("no id_token field in oauth2 token")
	}

	oidcConfig := &oidc.Config{
		ClientID: a.ClientID,
	}

	return a.Verifier(oidcConfig).Verify(ctx, rawIDToken)
}

// LoginURL returns the Auth0 authorization URL.
func LoginURL(state string) string {
	return cfg.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// ExchangeCode redeems the code for a token.
func ExchangeCode(code string) (*oauth2.Token, error) {
	return cfg.Exchange(context.Background(), code)
}
