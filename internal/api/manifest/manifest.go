package manifest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jcchavezs/nuro/internal/api"
	"github.com/jcchavezs/nuro/internal/http"
)

// GetConfigDigestFromManifest gets the digest of the config from the manifest
func GetConfigDigestFromManifest(ctx context.Context, registry string, insecure bool, name, reference string) (string, error) {
	var (
		digest string
		err    error
	)

	digest, err = GetConfigDigestFromManifestSingle(ctx, registry, insecure, name, reference)
	if err == nil {
		return digest, nil
	}

	digest, err = GetConfigDigestFromManifestList(ctx, registry, insecure, name, reference)
	if err == nil {
		return digest, nil
	}

	return digest, err
}

// GetConfigDigestFromManifestList gets the digest of the config from a list manifest
func GetConfigDigestFromManifestList(ctx context.Context, registry string, insecure bool, name, reference string) (string, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		fmt.Sprintf("%s://%s/v2/%s/manifests/%s", http.ResolveProtocol(insecure), registry, name, reference),
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.list.v2+json")
	req.Header.Add("Accept", "application/vnd.oci.image.index.v1+json")
	req.Header.Add("Accept", "application/vnd.oci.image.manifest.v1+json")
	req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	req.Header.Set("Accept-Encoding", "gzip")

	res, err := http.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("doing request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		var errRes api.ErrorResponse
		if err := json.NewDecoder(res.Body).Decode(&errRes); err != nil {
			return "", fmt.Errorf("decoding error response: %w", err)
		}

		return "", fmt.Errorf("unexpected status code %d: %w", res.StatusCode, errRes.Error())
	}

	switch res.Header.Get("Content-Type") {
	case ociImageV1ContentType:
		m := manifestList{}

		if err := json.NewDecoder(res.Body).Decode(&m); err != nil {
			return "", fmt.Errorf("decoding response: %w", err)
		}

		return m.Manifests[0].Digest, nil
	case manifestV2ContentType:
		m := manifest{}

		if err := json.NewDecoder(res.Body).Decode(&m); err != nil {
			return "", fmt.Errorf("decoding response: %w", err)
		}

		return m.Config.Digest, nil
	}
	return "", errors.New("unexpected content type")
}

type manifest struct {
	Config struct {
		Digest string `json:"digest"`
	} `json:"config"`
}

type manifestList struct {
	Manifests []struct {
		Digest string `json:"digest"`
	} `json:"manifests"`
}

const (
	manifestV2ContentType     = "application/vnd.docker.distribution.manifest.v2+json"
	manifestListV2ContentType = "application/vnd.docker.distribution.manifest.list.v2+json"
	ociImageV1ContentType     = "application/vnd.oci.image.index.v1+json"
)

// GetConfigDigestFromManifestSingle gets the digest of the config from a single manifest
func GetConfigDigestFromManifestSingle(ctx context.Context, registry string, insecure bool, name, reference string) (string, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		fmt.Sprintf("%s://%s/v2/%s/manifests/%s", http.ResolveProtocol(insecure), registry, name, reference),
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.list.v2+json")
	req.Header.Add("Accept", "application/vnd.oci.image.index.v1+json")
	req.Header.Add("Accept", "application/vnd.oci.image.manifest.v1+json")
	req.Header.Set("Accept-Encoding", "gzip")

	res, err := http.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("doing request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		var errRes api.ErrorResponse
		if err := json.NewDecoder(res.Body).Decode(&errRes); err != nil {
			return "", fmt.Errorf("decoding error response: %w", err)
		}

		return "", fmt.Errorf("unexpected status code %d: %w", res.StatusCode, errRes.Error())
	}

	switch res.Header.Get("Content-Type") {
	case manifestV2ContentType:
		m := manifest{}

		if err := json.NewDecoder(res.Body).Decode(&m); err != nil {
			return "", fmt.Errorf("decoding response: %w", err)
		}

		return m.Config.Digest, nil
	case manifestListV2ContentType:
		m := manifestList{}

		if err := json.NewDecoder(res.Body).Decode(&m); err != nil {
			return "", fmt.Errorf("decoding response: %w", err)
		}

		if len(m.Manifests) == 0 {
			return "", errors.New("no manifests found")
		}

		return GetConfigDigestFromManifestList(ctx, registry, insecure, name, m.Manifests[0].Digest)
	}

	return "", errors.New("unexpected content type")
}
