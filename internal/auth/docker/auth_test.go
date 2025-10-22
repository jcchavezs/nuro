package docker

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetToken(t *testing.T) {
	// Reset state before each test
	originalToken := cachedToken
	originalURL := baseAuthDockerURL
	originalClient := client
	defer func() {
		cachedToken = originalToken
		baseAuthDockerURL = originalURL
		client = originalClient
	}()

	t.Run("successful token retrieval", func(t *testing.T) {
		cachedToken = ""
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "registry.docker.io", r.URL.Query().Get("service"))
			require.Equal(t, "repository:library/nginx:pull", r.URL.Query().Get("scope"))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"token":"test-token-123"}`))
		}))
		defer server.Close()

		baseAuthDockerURL = server.URL
		client = server.Client()

		token, err := GetToken(context.Background(), "library/nginx")
		require.NoError(t, err)
		require.Equal(t, "test-token-123", token)
		require.Equal(t, "test-token-123", cachedToken)
	})

	t.Run("uses cached token", func(t *testing.T) {
		cachedToken = "cached-token-456"
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("should not make HTTP request when token is cached")
		}))
		defer server.Close()

		token, err := GetToken(context.Background(), "library/nginx")
		require.NoError(t, err)
		require.Equal(t, "cached-token-456", token)
	})

	t.Run("non-200 status code", func(t *testing.T) {
		cachedToken = ""
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
		}))
		defer server.Close()

		baseAuthDockerURL = server.URL
		client = server.Client()

		_, err := GetToken(context.Background(), "library/nginx")
		require.Error(t, err)
		require.Contains(t, err.Error(), "unexpected status code 401")
	})

	t.Run("invalid JSON response", func(t *testing.T) {
		cachedToken = ""
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`invalid json`))
		}))
		defer server.Close()

		baseAuthDockerURL = server.URL
		client = server.Client()

		_, err := GetToken(context.Background(), "library/nginx")
		require.Error(t, err)
		require.Contains(t, err.Error(), "decoding response")
	})
}
