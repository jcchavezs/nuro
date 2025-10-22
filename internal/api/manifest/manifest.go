package manifest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jcchavezs/nuro/internal/api"
	"github.com/jcchavezs/nuro/internal/http"
	"github.com/jcchavezs/nuro/internal/log"
	"go.uber.org/zap"
)

// GetConfigDigestFromManifest gets the digest of the config from the manifest
func GetConfigDigestFromManifest(ctx context.Context, registry string, insecure bool, name, reference string) (string, error) {
	log.Logger.Debug("Getting config digest from a manifest")
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
	log.Logger.Debug("Getting config digest from a manifest list")
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
	req.Header.Add("Accept", ociImageIndexV1ContentType)
	req.Header.Add("Accept", ociImageManifestV1ContentType)
	req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	req.Header.Set("Accept-Encoding", "gzip")

	res, err := http.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("doing request: %w", err)
	}
	defer res.Body.Close() //nolint

	if res.StatusCode != http.StatusOK {
		var errRes api.ErrorResponse
		if err := json.NewDecoder(res.Body).Decode(&errRes); err != nil {
			return "", fmt.Errorf("decoding error response: %w", err)
		}

		return "", fmt.Errorf("unexpected status code %d: %w", res.StatusCode, errRes.Error())
	}

	contentType := res.Header.Get("Content-Type")

	switch contentType {
	case ociImageIndexV1ContentType:
		m := manifestList{}

		if err := json.NewDecoder(res.Body).Decode(&m); err != nil {
			return "", fmt.Errorf("decoding %q response: %w", ociImageIndexV1ContentType, err)
		}
		//return "sha256:8e1939c6b02d37a2a16bf6e8dad40f95f2c8ede3595a4649729145fdf92efd34", nil
		return m.Manifests[0].Digest, nil
	case manifestV2ContentType, ociImageManifestV1ContentType:
		m := manifest{}

		if err := json.NewDecoder(res.Body).Decode(&m); err != nil {
			return "", fmt.Errorf("decoding %q response: %w", manifestV2ContentType, err)
		}

		return m.Config.Digest, nil
	default:
		log.Logger.Warn("Unexpected content type", zap.String("content-type", contentType))
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
	Annotations map[string]string `json:"annotations"`
}

const (
	manifestV2ContentType         = "application/vnd.docker.distribution.manifest.v2+json"
	manifestListV2ContentType     = "application/vnd.docker.distribution.manifest.list.v2+json"
	ociImageIndexV1ContentType    = "application/vnd.oci.image.index.v1+json"
	ociImageManifestV1ContentType = "application/vnd.oci.image.manifest.v1+json"
)

func GetAnnotationsFromManifestSingle(ctx context.Context, registry string, insecure bool, name, reference string) (map[string]string, bool, error) {
	log.Logger.Debug("Getting annotations from a manifest single")
	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		fmt.Sprintf("%s://%s/v2/%s/manifests/%s", http.ResolveProtocol(insecure), registry, name, reference),
		nil,
	)
	if err != nil {
		return nil, false, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Add("Accept", ociImageIndexV1ContentType)
	req.Header.Set("Accept-Encoding", "gzip")

	res, err := http.Client.Do(req)
	if err != nil {
		return nil, false, fmt.Errorf("doing request: %w", err)
	}
	defer res.Body.Close() //nolint

	if res.StatusCode != http.StatusOK {
		var errRes api.ErrorResponse

		if err := json.NewDecoder(res.Body).Decode(&errRes); err != nil {
			return nil, false, fmt.Errorf("decoding error response: %w", err)
		}

		log.Logger.Debug("OCI Image Manifest V1 not found")
		return nil, false, nil
	}

	if res.Header.Get("Content-Type") == ociImageIndexV1ContentType {
		m := manifestList{}

		if err := json.NewDecoder(res.Body).Decode(&m); err != nil {
			return nil, false, fmt.Errorf("decoding response: %w", err)
		}

		return m.Annotations, len(m.Annotations) > 0, nil
	}

	return nil, false, nil
}

// GetConfigDigestFromManifestSingle gets the digest of the config from a single manifest
func GetConfigDigestFromManifestSingle(ctx context.Context, registry string, insecure bool, name, reference string) (string, error) {
	log.Logger.Debug("Getting config digest from a manifest single")
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
	req.Header.Add("Accept", ociImageIndexV1ContentType)
	req.Header.Add("Accept", ociImageManifestV1ContentType)
	req.Header.Set("Accept-Encoding", "gzip")

	res, err := http.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("doing request: %w", err)
	}
	defer res.Body.Close() //nolint

	if res.StatusCode != http.StatusOK {
		var errRes api.ErrorResponse
		if err := json.NewDecoder(res.Body).Decode(&errRes); err != nil {
			return "", fmt.Errorf("decoding error response: %w", err)
		}

		return "", fmt.Errorf("unexpected status code %d: %w", res.StatusCode, errRes.Error())
	}
	contentType := res.Header.Get("Content-Type")
	switch contentType {
	case manifestV2ContentType:
		m := manifest{}

		if err := json.NewDecoder(res.Body).Decode(&m); err != nil {
			return "", fmt.Errorf("decoding response: %w", err)
		}

		return m.Config.Digest, nil
	case manifestListV2ContentType, ociImageIndexV1ContentType:
		m := manifestList{}

		if err := json.NewDecoder(res.Body).Decode(&m); err != nil {
			return "", fmt.Errorf("decoding response: %w", err)
		}

		if len(m.Manifests) == 0 {
			return "", errors.New("no manifests found")
		}

		return GetConfigDigestFromManifestList(ctx, registry, insecure, name, m.Manifests[0].Digest)
	default:
		log.Logger.Warn("Unexpected content type", zap.String("content-type", contentType))
	}

	return "", errors.New("unexpected content type")
}
