package application

import (
	"context"
	"log"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type GoogleAuth struct {
	verifier *oidc.IDTokenVerifier
	config   oauth2.Config
}

func NewGoogleAuth(clientID, clientSecret, providerURL, redirectURL string) *GoogleAuth {
	ctx := context.Background()

	provider, err := oidc.NewProvider(ctx, providerURL)
	if err != nil {
		log.Fatal(err)
	}
	oidcConfig := &oidc.Config{
		ClientID: clientID,
	}
	verifier := provider.Verifier(oidcConfig)

	config := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  redirectURL,
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	return &GoogleAuth{
		verifier: verifier,
		config:   config,
	}
}

func (g *GoogleAuth) Verify(ctx context.Context, code string) (*oidc.IDToken, error) {
	oauth2Token, err := g.config.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		return nil, err
	}
	idToken, err := g.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, err
	}

	return idToken, nil
}

func (g *GoogleAuth) AuthCodeURL(state string, nonce string) string {
	return g.config.AuthCodeURL(state, oidc.Nonce(nonce))
}
