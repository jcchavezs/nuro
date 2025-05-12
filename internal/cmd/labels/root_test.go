package labels

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetLabelsFromBlob(t *testing.T) {
	tests := []struct {
		name           string
		nameParam      string
		digest         string
		mockResponse   string
		mockStatusCode int
		expectedLabels json.RawMessage
		expectErr      bool
	}{
		{
			name:           "valid response with labels",
			nameParam:      "library/nginx",
			digest:         "sha256:abc123",
			mockResponse:   `{"config": {"labels": {"key1": "value1", "key2": "value2"}}}`,
			mockStatusCode: http.StatusOK,
			expectedLabels: json.RawMessage(`{"key1": "value1", "key2": "value2"}`),
			expectErr:      false,
		},
		{
			name:           "valid response with no labels",
			nameParam:      "library/nginx",
			digest:         "sha256:abc123",
			mockResponse:   `{"config": {"labels": {}}}`,
			mockStatusCode: http.StatusOK,
			expectedLabels: json.RawMessage(`{}`),
			expectErr:      false,
		},
		{
			name:           "error response from server",
			nameParam:      "library/nginx",
			digest:         "sha256:abc123",
			mockResponse:   `{"errors": [{"message": "not found"}]}`,
			mockStatusCode: http.StatusNotFound,
			expectedLabels: nil,
			expectErr:      true,
		},
		{
			name:           "invalid JSON response",
			nameParam:      "library/nginx",
			digest:         "sha256:abc123",
			mockResponse:   `invalid-json`,
			mockStatusCode: http.StatusOK,
			expectedLabels: nil,
			expectErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock HTTP server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, "/v2/"+tt.nameParam+"/blobs/"+tt.digest, r.URL.Path)
				w.WriteHeader(tt.mockStatusCode)
				_, _ = w.Write([]byte(tt.mockResponse))
			}))
			defer server.Close()

			// Replace the registry with the mock server URL
			registry := server.URL[len("http://"):]

			// Call the function
			labels, err := getLabelsFromBlob(context.Background(), registry, true, tt.nameParam, tt.digest)

			// Validate results
			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedLabels, labels)
			}
		})
	}
}
