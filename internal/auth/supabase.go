package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"tukychat/internal/config"
	"tukychat/internal/models"
)

type Client struct {
	baseURL    string
	anonKey    string
	httpClient *http.Client
}

func NewClient(cfg config.Config) *Client {
	return &Client{
		baseURL: strings.TrimRight(cfg.SupabaseURL, "/"),
		anonKey: cfg.SupabaseAnonKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) GetUser(accessToken string) (*models.SupabaseUser, int, error) {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+"/auth/v1/user", nil)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("crear request auth: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("apikey", c.anonKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, http.StatusBadGateway, fmt.Errorf("consultar supabase auth: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode, nil
	}

	var user models.SupabaseUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("decodear usuario auth: %w", err)
	}

	return &user, http.StatusOK, nil
}