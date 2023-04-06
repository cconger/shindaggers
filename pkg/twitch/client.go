package twitch

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
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

// debugBody can wrap a resp.Body reader for debugging a payload before it enters a JSON decoder
func debugBody(r io.Reader) io.Reader {
	return io.TeeReader(r, os.Stdout)
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

func (c *Client) GetAppToken(ctx context.Context) (*GetTokenResponse, error) {
	params := url.Values{}
	params.Add("client_id", c.ClientID)
	params.Add("client_secret", c.ClientSecret)
	params.Add("grant_type", "client_credentials")

	req, err := http.NewRequest(http.MethodPost, "https://id.twitch.tv/oauth2/token", strings.NewReader(params.Encode()))
	if err != nil {
		return nil, fmt.Errorf("creating oauth request: %w", err)
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("post oauth request: %w", err)
	}
	defer resp.Body.Close()

	var tokenResp GetTokenResponse
	err = json.NewDecoder(resp.Body).Decode(&tokenResp)
	if err != nil {
		return nil, fmt.Errorf("deserializing oauth: %w", err)
	}

	return &tokenResp, nil
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
	Data []*TwitchUser `json:"data"`
}

func (c *Client) GetUser(ctx context.Context, token string) (*TwitchUser, error) {
	users, err := c.GetUsersByLogin(ctx, token)
	return users[0], err
}

func (c *Client) GetUsersByLogin(ctx context.Context, token string, login ...string) ([]*TwitchUser, error) {
	u, err := url.Parse("https://api.twitch.tv/helix/users")
	if err != nil {
		return nil, err
	}

	params := url.Values{
		"login": login,
	}
	u.RawQuery = params.Encode()

	r, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
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

	return payload.Data, nil
}
