// Package oauthhelper provide an OAuth2 authentication and token management helper
package oauthhelper

import (
	"context"
	base642 "encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
)

// AuthenticateFunc defines the signature of the authentication function used
type AuthenticateFunc func(url string) (code string, err error)

// Auth defines the authentication parameters
type Auth struct {
	// Token holds the token that should be used for authentication (optional)
	// if the token is nil the callback func Authenticate will be called and after Authorization this token will be set
	// Store (and restore prior use) this token to avoid further authorization calls
	Token *oauth2.Token
	// ClientID  from https://console.developers.google.com/project/<your-project-id>/apiui/credential
	ClientID     string
	ClientSecret string
	Authenticate AuthenticateFunc
}

// NewHTTPClient instantiates a new authentication client
func (auth *Auth) NewHTTPClient(ctx context.Context, scopes ...string) (*http.Client, error) {
	// If no scope has been specified, it shall only be the drive API one
	if len(scopes) == 0 {
		scopes = []string{"https://www.googleapis.com/auth/drive"}
	}

	config := &oauth2.Config{
		Scopes:      scopes,
		RedirectURL: "http://localhost",
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: "https://accounts.google.com/o/oauth2/token",
		},
		ClientID:     auth.ClientID,
		ClientSecret: auth.ClientSecret,
	}

	if auth.Token == nil {
		var err error

		auth.Token, err = auth.getTokenFromWeb(config)
		if err != nil {
			return nil, err
		}
	}

	return config.Client(ctx, auth.Token), nil
}

func (auth *Auth) getTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)

	code, err := auth.Authenticate(authURL)
	if err != nil {
		return nil, fmt.Errorf("authenticate error: %w", err)
	}

	tok, err := config.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve token from web: %w", err)
	}

	return tok, nil
}

// LoadTokenFromFile loads an OAuth2 token from a JSON file
func LoadTokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(filepath.Clean(file))
	if err != nil {
		return nil, fmt.Errorf("couldn't open token file: %w", err)
	}

	var token oauth2.Token
	err = json.NewDecoder(f).Decode(&token)

	if errClose := f.Close(); errClose != nil {
		return nil, fmt.Errorf("couldn't close token file: %w", errClose)
	}

	return &token, err
}

// StoreTokenToFile stores an OAuth2 token to a JSON file
func StoreTokenToFile(file string, token *oauth2.Token) error {
	f, err := os.Create(file)
	if err != nil {
		return fmt.Errorf("couldn't open token file: %w", err)
	}

	err = json.NewEncoder(f).Encode(token)

	if errClose := f.Close(); errClose != nil {
		return fmt.Errorf("couldn't close token file: %w", errClose)
	}

	return err
}

// GetTokenBase64 returns the Base64 representation of JSON token
func GetTokenBase64(token *oauth2.Token) (string, error) {
	jb, err := json.Marshal(token)
	if err != nil {
		return "", err
	}

	return base642.URLEncoding.EncodeToString(jb), nil
}
