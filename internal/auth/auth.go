package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jcchavezs/nuro/internal/auth/docker"
	"github.com/jcchavezs/nuro/internal/image"
	"github.com/jcchavezs/nuro/internal/log"
	"github.com/jdx/go-netrc"
	"go.uber.org/zap"
)

type ImageMetadata struct {
	Registry string
	Name     string
}

type imageMetadataKey struct{}

var ctxKey imageMetadataKey

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
	return context.WithValue(ctx, ctxKey, metadata)
}

type authRoundTripper struct {
	http.RoundTripper
}

func (rt authRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if metadata, ok := req.Context().Value(ctxKey).(ImageMetadata); ok {
		if metadata.Registry == image.DockerRegistry {
			// we should only inject credentials when the host is docker registry, if
			// it is a redirection e.g. CDN we shouldn't inject the credentials.
			if req.URL.Host == image.DockerRegistry {
				if token, err := docker.GetToken(req.Context(), metadata.Name); err != nil {
					return nil, fmt.Errorf("authenticating in docker registry: %w", err)
				} else if token != "" {
					log.Logger.Debug("Setting docker authorization")
					req.Header.Set("Authorization", "Bearer "+token)
				} else {
					log.Logger.Debug("Empty docker token")
				}
			}
		} else if netRC != nil {
			if metadata.Registry != req.URL.Host {
				log.Logger.Warn(
					"request URL host and registry host aren't the same",
					zap.String("image_registry", metadata.Registry),
					zap.String("request_url_host", req.URL.Host),
				)
			}
			// Check if we have a netrc entry for the registry
			if m := netRC.Machine(metadata.Registry); m != nil {
				log.Logger.Debug("Setting netrc authorization")
				req.Header.Set("Authorization", "Bearer "+m.Get("password"))
			} else {
				log.Logger.Debug("Netrc authorization not found")
			}
		}
	}

	return rt.RoundTripper.RoundTrip(req)
}

func WrapRoundTripper(t http.RoundTripper) http.RoundTripper {
	return authRoundTripper{t}
}
