package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/jcchavezs/nuro/internal/log"
)

var client = &http.Client{
	Transport: log.WrapRoundTripper(http.DefaultTransport),
}

var cachedToken string

func GetToken(ctx context.Context, image string) (string, error) {
	if cachedToken != "" {
		log.Logger.Debug("Using cached docker token")
		return cachedToken, nil
	}

	resp, err := client.Get("https://auth.docker.io/token?service=registry.docker.io&scope=repository:" + image + ":pull")
	if err != nil {
		return "", fmt.Errorf("doing request: %w", err)
	}
	defer resp.Body.Close() //nolint

	if resp.StatusCode != http.StatusOK {
		io.Copy(io.Discard, resp.Body)
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
