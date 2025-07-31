package blob

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jcchavezs/nuro/internal/api"
	"github.com/jcchavezs/nuro/internal/http"
)

type ConfigBlob struct {
	Config struct {
		Labels map[string]string `json:"labels"`
	} `json:"config"`
	Annotations map[string]string `json:"annotations"`
	Created     time.Time         `json:"created"`
}

// GetConfigBlob gets the config blob using a digest
func GetConfigBlob(ctx context.Context, registry string, insecure bool, name, digest string) (*ConfigBlob, error) {
	req, err := http.NewRequestWithContext(
		ctx, "GET",
		fmt.Sprintf("%s://%s/v2/%s/blobs/%s", http.ResolveProtocol(insecure), registry, name, digest),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	res, err := http.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("doing request: %w", err)
	}
	defer res.Body.Close() //nolint

	if res.StatusCode != http.StatusOK {
		var errRes api.ErrorResponse
		if err := json.NewDecoder(res.Body).Decode(&errRes); err != nil {
			return nil, fmt.Errorf("decoding error response: %w", err)
		}

		return nil, fmt.Errorf("unexpected status code %d: %w", res.StatusCode, errRes.Error())
	}

	c := &ConfigBlob{}

	if err := json.NewDecoder(res.Body).Decode(c); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return c, nil
}
