package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jcchavezs/nuro/internal/auth/docker"
	"github.com/jcchavezs/nuro/internal/image"
	"github.com/jdx/go-netrc"
)

type ImageMetadata struct {
	Registry string
	Name     string
}

var imageMetadataKey = struct{}{}

var netRC *netrc.Netrc

func LoadNetRCFile(ctx context.Context, netRCFile string) error {
	var err error
	netRC, err = netrc.Parse(netRCFile)
	return err
}

func LoadNetRC(ctx context.Context, netRCContents string) error {
	var err error
	netRC, err = netrc.ParseString(netRCContents)
	return err
}

func InjectImageMetadata(ctx context.Context, metadata ImageMetadata) context.Context {
	return context.WithValue(ctx, imageMetadataKey, metadata)
}

type authRoundTripper struct {
	http.RoundTripper
}

func (rt authRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if metadata, ok := req.Context().Value(imageMetadataKey).(ImageMetadata); ok {
		if metadata.Registry == image.DockerRegistry {
			if token, err := docker.GetToken(req.Context(), metadata.Name); err != nil {
				return nil, fmt.Errorf("authenticating in docker registry: %w", err)
			} else {
				req.Header.Set("Authorization", "Bearer "+token)
			}
		} else if netRC != nil {
			// Check if we have a netrc entry for the registry
			if m := netRC.Machine(metadata.Registry); m != nil {
				req.Header.Set("Authorization", "Bearer "+m.Get("password"))
			}
		}
	}

	return rt.RoundTripper.RoundTrip(req)
}

func WrapRoundTripper(t http.RoundTripper) http.RoundTripper {
	return authRoundTripper{t}
}
