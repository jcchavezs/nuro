package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

var cachedToken string

func GetToken(ctx context.Context, image string) (string, error) {
	if cachedToken != "" {
		return cachedToken, nil
	}

	resp, err := http.DefaultClient.Get("https://auth.docker.io/token?service=registry.docker.io&scope=repository:" + image + ":pull")
	if err != nil {
		return "", fmt.Errorf("doing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	var result struct {
		Token string `json:"token"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decoding response: %w", err)
	}

	cachedToken = result.Token
	return cachedToken, nil
}
