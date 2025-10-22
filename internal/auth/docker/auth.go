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

var (
	cachedToken       string
	baseAuthDockerURL = "https://auth.docker.io"
)

func GetToken(ctx context.Context, image string) (string, error) {
	if cachedToken != "" {
		log.Logger.Debug("Using cached docker token")
		return cachedToken, nil
	}

	tokenURL := fmt.Sprintf("%s/token?service=registry.docker.io&scope=repository:%s:pull", baseAuthDockerURL, image)
	resp, err := client.Get(tokenURL)
	if err != nil {
		return "", fmt.Errorf("doing request: %w", err)
	}
	defer resp.Body.Close() //nolint

	if resp.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, resp.Body)
		// TODO: deal with error response
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
