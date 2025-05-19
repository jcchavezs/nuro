package manifest

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetConfigDigestFromManifestSingle(t *testing.T) {
	tests := []struct {
		name            string
		nameParam       string
		reference       string
		mockResponse    string
		mockStatusCode  int
		mockContentType string
		expectedDigest  string
		expectErr       bool
	}{
		{
			name:            "valid manifest v2 response",
			nameParam:       "library/nginx",
			reference:       "latest",
			mockResponse:    `{"config": {"digest": "sha256:abc123"}}`,
			mockStatusCode:  http.StatusOK,
			mockContentType: manifestV2ContentType,
			expectedDigest:  "sha256:abc123",
			expectErr:       false,
		},
		{
			name:            "unexpected content type",
			nameParam:       "library/nginx",
			reference:       "latest",
			mockResponse:    `{"config": {"digest": "sha256:abc123"}}`,
			mockStatusCode:  http.StatusOK,
			mockContentType: "application/unknown",
			expectedDigest:  "",
			expectErr:       true,
		},
		{
			name:            "error response from server",
			nameParam:       "library/nginx",
			reference:       "latest",
			mockResponse:    `{"errors": [{"message": "not found"}]}`,
			mockStatusCode:  http.StatusNotFound,
			mockContentType: "application/json",
			expectedDigest:  "",
			expectErr:       true,
		},
		{
			name:            "invalid JSON response",
			nameParam:       "library/nginx",
			reference:       "latest",
			mockResponse:    `invalid-json`,
			mockStatusCode:  http.StatusOK,
			mockContentType: manifestV2ContentType,
			expectedDigest:  "",
			expectErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock HTTP server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, "/v2/"+tt.nameParam+"/manifests/"+tt.reference, r.URL.Path)
				w.Header().Set("Content-Type", tt.mockContentType)
				w.WriteHeader(tt.mockStatusCode)
				_, _ = w.Write([]byte(tt.mockResponse))
			}))
			defer server.Close()

			// Replace the registry with the mock server URL
			registry := server.URL[len("http://"):]

			// Call the function
			digest, err := GetConfigDigestFromManifestSingle(context.Background(), registry, true, tt.nameParam, tt.reference)

			// Validate results
			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedDigest, digest)
			}
		})
	}
}
