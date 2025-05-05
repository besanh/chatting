package oauth2

import (
	"context"

	"github.com/besanh/chatting/config"
	"golang.org/x/oauth2"
)

type (
	IOAuth2 interface {
		GetClient() *oauth2.Config
		AuthCodeUrl(state, verifier string) string
		Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error)
	}

	OAuth2 struct {
		Config config.Oauth2
	}
)

func NewOAuth2(config config.Oauth2) IOAuth2 {
	return &OAuth2{
		Config: config,
	}
}

func (o *OAuth2) GetClient() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     o.Config.Google.ClientId,
		ClientSecret: o.Config.Google.ClientSecret,
		Scopes:       o.Config.Google.Scope,
		Endpoint:     o.Config.Google.Endpoint,
		RedirectURL:  o.Config.Google.RedirectUrl,
	}
}

func (o *OAuth2) AuthCodeUrl(state, verifier string) string {
	return o.GetClient().AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.S256ChallengeOption(verifier))
}

func (o *OAuth2) Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	return o.GetClient().Exchange(ctx, code, opts...)
}
