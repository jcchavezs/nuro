package created

import (
	"testing"
	"time"

	"github.com/jcchavezs/nuro/internal/api/blob"
	"github.com/stretchr/testify/require"
)

func TestResolveDateFromConfig(t *testing.T) {
	tests := []struct {
		name         string
		cfg          *blob.ConfigBlob
		expectedDate string
		expectedOk   bool
	}{
		{
			name: "created field is set",
			cfg: &blob.ConfigBlob{
				Created: time.Date(2023, 10, 1, 12, 0, 0, 0, time.UTC),
			},
			expectedDate: "2023-10-01T12:00:00Z",
			expectedOk:   true,
		},
		{
			name: "created field is zero, annotations contain creation date",
			cfg: &blob.ConfigBlob{
				Annotations: map[string]string{
					"org.opencontainers.image.created": "2023-10-01T12:00:00Z",
				},
			},
			expectedDate: "2023-10-01T12:00:00Z",
			expectedOk:   true,
		},
		{
			name: "created field is zero, labels contain creation date",
			cfg: &blob.ConfigBlob{
				Config: struct {
					Labels map[string]string `json:"labels"`
				}{
					Labels: map[string]string{
						"org.opencontainers.image.created": "2023-10-01T12:00:00Z",
					},
				},
			},
			expectedDate: "2023-10-01T12:00:00Z",
			expectedOk:   true,
		},
		{
			name:         "created field is zero, no annotations or labels",
			cfg:          &blob.ConfigBlob{},
			expectedDate: "",
			expectedOk:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			date, ok := resolveDateFromConfig(tt.cfg)
			require.Equal(t, tt.expectedOk, ok)
			require.Equal(t, tt.expectedDate, date)
		})
	}
}
