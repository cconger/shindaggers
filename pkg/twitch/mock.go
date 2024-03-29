package twitch

import (
	"context"
	"fmt"
)

type MockClient struct{}

func (m *MockClient) OAuthGetToken(ctx context.Context, code string, redirectURI string) (*GetTokenResponse, error) {
	return nil, fmt.Errorf("oauth not supported on dev env")
}

func (m *MockClient) GetUser(ctx context.Context) (*TwitchUser, error) {
	return nil, fmt.Errorf("cannot lookup users with mock twitch client")
}

func (m *MockClient) GetUsersByID(ctx context.Context, users ...string) ([]*TwitchUser, error) {
	return nil, fmt.Errorf("cannot lookup users with mock twitch client")
}

func (m *MockClient) UserClient(ua *UserAuth) UserClient {
	return m
}
