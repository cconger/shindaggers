package twitch

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type Client struct {
	Client       *http.Client
	ClientID     string
	ClientSecret string
}

func NewClient(clientID string, clientSecret string) (*Client, error) {
	if clientID == "" {
		return nil, fmt.Errorf("client id cannot be empty")
	}

	if clientSecret == "" {
		return nil, fmt.Errorf("client secret cannot be empty")
	}

	return &Client{
		// TODO: More sane default
		Client:       &http.Client{},
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}, nil
}

func (c *Client) authHeaders(r *http.Request, token string) *http.Request {
	r.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	r.Header.Add("Client-Id", c.ClientID)
	return r
}

type GetTokenResponse struct {
	AccessToken  string   `json:"access_token"`
	ExpiresIn    int64    `json:"expires_in"`
	RefreshToken string   `json:"refresh_token"`
	Scope        []string `json:"scope"`
	TokenType    string   `json:"token_type"`
}

func (c *Client) OAuthGetToken(ctx context.Context, code string, redirectURI string) (*GetTokenResponse, error) {
	// POST to https://id.twitch.tv/oauth2/token

	payload := url.Values{
		"client_id":     []string{c.ClientID},
		"client_secret": []string{c.ClientSecret},
		"code":          []string{code},
		"grant_type":    []string{"authorization_code"},
		"redirect_uri":  []string{redirectURI},
	}

	r, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://id.twitch.tv/oauth2/token", strings.NewReader(payload.Encode()))
	if err != nil {
		return nil, err
	}

	resp, err := c.Client.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response GetTokenResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

type TwitchUser struct {
	ID          string `json:"id"`
	Login       string `json:"login"`
	DisplayName string `json:"display_name"`
}

type TwitchUserPayload struct {
	Data []TwitchUser `json:"data"`
}

func (c *Client) GetUser(ctx context.Context, token string) (*TwitchUser, error) {
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.twitch.tv/helix/users", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.Client.Do(c.authHeaders(r, token))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var payload TwitchUserPayload
	err = json.NewDecoder(resp.Body).Decode(&payload)
	if err != nil {
		return nil, err
	}

	if len(payload.Data) < 1 {
		return nil, fmt.Errorf("no results")
	}

	return &payload.Data[0], nil
}
