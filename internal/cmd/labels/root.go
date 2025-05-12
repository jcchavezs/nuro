package labels

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jcchavezs/nuro/internal/http"
	"github.com/spf13/cobra"
)

type errorResponse struct {
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func (r errorResponse) Error() error {
	if len(r.Errors) == 0 {
		return nil
	}

	errs := make([]error, 0, len(r.Errors))
	for _, e := range r.Errors {
		errs = append(errs, errors.New(e.Message))
	}

	return errors.Join(errs...)
}

func getConfigDigestFromManifest(ctx context.Context, registry, name, reference string) (string, error) {
	var (
		digest string
		err    error
	)

	if digest, err = getConfigDigestFromManifestSingle(ctx, registry, name, reference); err == nil {
		return digest, nil
	}

	if digest, err = getConfigDigestFromManifestList(ctx, registry, name, reference); err == nil {
		return digest, nil
	}

	return digest, err
}

func getConfigDigestFromManifestList(ctx context.Context, registry, name, reference string) (string, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		fmt.Sprintf("https://%s/v2/%s/manifests/%s", registry, name, reference),
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
		var errRes errorResponse
		if err := json.NewDecoder(res.Body).Decode(&errRes); err != nil {
			return "", fmt.Errorf("decoding error response: %w", err)
		}

		return "", fmt.Errorf("unexpected status code %d: %w", res.StatusCode, errRes.Error())
	}

	type manifest struct {
		Config struct {
			Digest string `json:"digest"`
		} `json:"config"`
	}

	m := manifest{}

	if err := json.NewDecoder(res.Body).Decode(&m); err != nil {
		return "", fmt.Errorf("decoding response: %w", err)
	}

	return m.Config.Digest, nil
}

func getConfigDigestFromManifestSingle(ctx context.Context, registry, name, reference string) (string, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		fmt.Sprintf("https://%s/v2/%s/manifests/%s", registry, name, reference),
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
		var errRes errorResponse
		if err := json.NewDecoder(res.Body).Decode(&errRes); err != nil {
			return "", fmt.Errorf("decoding error response: %w", err)
		}

		return "", fmt.Errorf("unexpected status code %d: %w", res.StatusCode, errRes.Error())
	}

	type manifestList struct {
		Manifests []struct {
			Digest string `json:"digest"`
		} `json:"manifests"`
	}

	m := manifestList{}

	if err := json.NewDecoder(res.Body).Decode(&m); err != nil {
		return "", fmt.Errorf("decoding response: %w", err)
	}

	return getConfigDigestFromManifestList(ctx, registry, name, m.Manifests[0].Digest)
}

type configBlob struct {
	Config struct {
		Labels json.RawMessage `json:"labels"`
	} `json:"config"`
}

func getLabelsFromBlob(ctx context.Context, registry, name, digest string) (json.RawMessage, error) {
	req, err := http.NewRequestWithContext(
		ctx, "GET",
		fmt.Sprintf("https://%s/v2/%s/blobs/%s", registry, name, digest),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	res, err := http.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("doing request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		var errRes errorResponse
		if err := json.NewDecoder(res.Body).Decode(&errRes); err != nil {
			return nil, fmt.Errorf("decoding error response: %w", err)
		}

		return nil, fmt.Errorf("unexpected status code %d: %w", res.StatusCode, errRes.Error())
	}

	c := configBlob{}

	if err := json.NewDecoder(res.Body).Decode(&c); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return c.Config.Labels, nil
}

func init() {
	RootCmd.AddCommand(ListCmd)
}

var RootCmd = &cobra.Command{
	Use:   "labels",
	Short: "Labels related commands",
}
