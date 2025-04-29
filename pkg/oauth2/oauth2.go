package oauth

import (
	"context"

	"golang.org/x/oauth2"
)

type (
	IOAuth2 interface {
		GetClient() *oauth2.Config
		AuthCodeUrl(state, verifier string) string
		Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error)
	}

	OAuth2 struct {
		Config OAuth2Config
	}

	OAuth2Config struct {
		ClientId     string
		ClientSecret string
		Scopes       []string
		Endpoint     oauth2.Endpoint
		Redirect     string
	}
)

func NewOAuth2(config OAuth2Config) IOAuth2 {
	return &OAuth2{Config: config}
}

func (o *OAuth2) GetClient() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     o.Config.ClientId,
		ClientSecret: o.Config.ClientSecret,
		Scopes:       o.Config.Scopes,
		Endpoint:     o.Config.Endpoint,
		RedirectURL:  o.Config.Redirect,
	}
}

func (o *OAuth2) AuthCodeUrl(state, verifier string) string {
	return o.GetClient().AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.S256ChallengeOption(verifier))
}

func (o *OAuth2) Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	return o.GetClient().Exchange(ctx, code, opts...)
}
