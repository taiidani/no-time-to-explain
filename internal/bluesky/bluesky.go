package bluesky

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type BlueskyClient struct {
	BaseURL    string
	AuthToken  string
	HttpClient *http.Client
}

// NewBlueskyClient creates a new client with the given auth token
func NewBlueskyClient() *BlueskyClient {
	return &BlueskyClient{
		BaseURL:    "https://public.api.bsky.app", // Base URL for Bluesky API
		HttpClient: &http.Client{},
	}
}

// UserData contains basic identifying information about the given user.
type UserData struct {
	DID         string `json:"did"`
	Handle      string `json:"handle"`
	DisplayName string `json:"displayName"`
	Avatar      string `json:"avatar"`
	Description string `json:"description"`
	Banner      string `json:"banner"`
}

// GetUser fetches user data by handle
func (c *BlueskyClient) GetUser(handle string) (*UserData, error) {
	params := url.Values{}
	params.Add("actor", handle)
	url := fmt.Sprintf("%s/xrpc/app.bsky.actor.getProfile?%s", c.BaseURL, params.Encode())

	resp, err := c.HttpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %s", string(bodyBytes))
	}

	var userData UserData
	if err := json.NewDecoder(resp.Body).Decode(&userData); err != nil {
		return nil, err
	}

	return &userData, nil
}
