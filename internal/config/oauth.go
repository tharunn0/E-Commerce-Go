package config

import "golang.org/x/oauth2"

func NewOAuthConfig(clientID, clientSecret, redirectUrl string) *oauth2.Config {
	return &oauth2.Config{
		ClientSecret: clientSecret,
		ClientID:     clientID,
		RedirectURL:  redirectUrl,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     oauth2.Endpoint{AuthURL: "https://accounts.google.com/o/oauth2/auth", TokenURL: "https://oauth2.googleapis.com/token"},
	}
}
