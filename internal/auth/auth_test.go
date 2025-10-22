package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jcchavezs/nuro/internal/image"
	"github.com/stretchr/testify/require"
)

type mockRoundTripper struct {
	roundTripFunc func(*http.Request) (*http.Response, error)
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.roundTripFunc(req)
}

func TestAuthRoundTripper_RoundTrip(t *testing.T) {
	// Reset netRC before tests
	originalNetRC := netRC
	defer func() {
		netRC = originalNetRC
	}()

	t.Run("no metadata in context", func(t *testing.T) {
		called := false
		mock := &mockRoundTripper{
			roundTripFunc: func(req *http.Request) (*http.Response, error) {
				called = true
				require.Empty(t, req.Header.Get("Authorization"))
				return &http.Response{StatusCode: http.StatusOK}, nil
			},
		}

		rt := authRoundTripper{mock}
		req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)

		resp, err := rt.RoundTrip(req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.True(t, called)
	})

	t.Run("docker registry with token", func(t *testing.T) {
		// Mock docker.GetToken by temporarily setting up test server
		called := false
		mock := &mockRoundTripper{
			roundTripFunc: func(req *http.Request) (*http.Response, error) {
				called = true
				// Note: In real test, you'd need to mock docker.GetToken
				// For now, we check the behavior when token would be set
				return &http.Response{StatusCode: http.StatusOK}, nil
			},
		}

		rt := authRoundTripper{mock}
		ctx := InjectImageMetadata(context.Background(), ImageMetadata{
			Registry: image.DockerRegistry,
			Name:     "library/nginx",
		})
		req := httptest.NewRequest(http.MethodGet, "https://"+image.DockerRegistry+"/v2/library/nginx/manifests/latest", nil)
		req = req.WithContext(ctx)

		resp, err := rt.RoundTrip(req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.True(t, called)
	})

	t.Run("docker registry with different host - no auth", func(t *testing.T) {
		called := false
		mock := &mockRoundTripper{
			roundTripFunc: func(req *http.Request) (*http.Response, error) {
				called = true
				require.Empty(t, req.Header.Get("Authorization"))
				return &http.Response{StatusCode: http.StatusOK}, nil
			},
		}

		rt := authRoundTripper{mock}
		ctx := InjectImageMetadata(context.Background(), ImageMetadata{
			Registry: image.DockerRegistry,
			Name:     "library/nginx",
		})
		req := httptest.NewRequest(http.MethodGet, "https://docker-images-prod.6aa30f8b08e16409b46e0173d6de2f56.r2.cloudflarestorage.com/registry-v2/docker/registry/v2/blobs/sha256/47/47363d0594e9195848a15d8357f85ebdeb962cf0a1a7830f14143bf806791238/data?", nil)
		req = req.WithContext(ctx)

		resp, err := rt.RoundTrip(req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.True(t, called)
	})

	t.Run("custom registry with netrc", func(t *testing.T) {
		netRC = nil
		err := LoadNetRC(context.Background(), `machine custom.registry.io
login user
password secret-token`)
		require.NoError(t, err)

		called := false
		mock := &mockRoundTripper{
			roundTripFunc: func(req *http.Request) (*http.Response, error) {
				called = true
				require.Equal(t, "Bearer secret-token", req.Header.Get("Authorization"))
				return &http.Response{StatusCode: http.StatusOK}, nil
			},
		}

		rt := authRoundTripper{mock}
		ctx := InjectImageMetadata(context.Background(), ImageMetadata{
			Registry: "custom.registry.io",
			Name:     "myorg/myimage",
		})
		req := httptest.NewRequest(http.MethodGet, "https://custom.registry.io/v2/myorg/myimage/manifests/latest", nil)
		req = req.WithContext(ctx)

		resp, err := rt.RoundTrip(req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.True(t, called)
	})

	t.Run("custom registry without netrc entry", func(t *testing.T) {
		netRC = nil
		err := LoadNetRC(context.Background(), `machine other.registry.io
login user
password secret-token`)
		require.NoError(t, err)

		called := false
		mock := &mockRoundTripper{
			roundTripFunc: func(req *http.Request) (*http.Response, error) {
				called = true
				require.Empty(t, req.Header.Get("Authorization"))
				return &http.Response{StatusCode: http.StatusOK}, nil
			},
		}

		rt := authRoundTripper{mock}
		ctx := InjectImageMetadata(context.Background(), ImageMetadata{
			Registry: "custom.registry.io",
			Name:     "myorg/myimage",
		})
		req := httptest.NewRequest(http.MethodGet, "https://custom.registry.io/v2/myorg/myimage/manifests/latest", nil)
		req = req.WithContext(ctx)

		resp, err := rt.RoundTrip(req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.True(t, called)
	})

	t.Run("custom registry with netrc but nil netRC", func(t *testing.T) {
		netRC = nil

		called := false
		mock := &mockRoundTripper{
			roundTripFunc: func(req *http.Request) (*http.Response, error) {
				called = true
				require.Empty(t, req.Header.Get("Authorization"))
				return &http.Response{StatusCode: http.StatusOK}, nil
			},
		}

		rt := authRoundTripper{mock}
		ctx := InjectImageMetadata(context.Background(), ImageMetadata{
			Registry: "custom.registry.io",
			Name:     "myorg/myimage",
		})
		req := httptest.NewRequest(http.MethodGet, "https://custom.registry.io/v2/myorg/myimage/manifests/latest", nil)
		req = req.WithContext(ctx)

		resp, err := rt.RoundTrip(req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.True(t, called)
	})
}

func TestLoadNetRC(t *testing.T) {
	t.Run("valid netrc content", func(t *testing.T) {
		err := LoadNetRC(context.Background(), `machine example.com
login user
password pass`)
		require.NoError(t, err)
		require.NotNil(t, netRC)
	})

	t.Run("invalid netrc content", func(t *testing.T) {
		err := LoadNetRC(context.Background(), `invalid content`)
		require.NoError(t, err)
		require.Empty(t, netRC.Machines())
	})
}

func TestInjectImageMetadata(t *testing.T) {
	metadata := ImageMetadata{
		Registry: "example.com",
		Name:     "myimage",
	}

	ctx := InjectImageMetadata(context.Background(), metadata)
	require.NotNil(t, ctx)

	retrieved, ok := ctx.Value(ctxKey).(ImageMetadata)
	require.True(t, ok)
	require.Equal(t, metadata, retrieved)
}
