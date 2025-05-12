package image

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseImage(t *testing.T) {
	tests := []struct {
		name           string
		image          string
		expectedReg    string
		expectedName   string
		expectedTag    string
		expectedDigest string
		expectErr      bool
	}{
		{
			name:           "well known image",
			image:          "nginx:latest",
			expectedReg:    "registry-1.docker.io",
			expectedName:   "library/nginx",
			expectedTag:    "latest",
			expectedDigest: "",
			expectErr:      false,
		},
		{
			name:           "Valid image with default registry",
			image:          "library/nginx:latest",
			expectedReg:    "registry-1.docker.io",
			expectedName:   "library/nginx",
			expectedTag:    "latest",
			expectedDigest: "",
			expectErr:      false,
		},
		{
			name:           "Valid image with custom registry",
			image:          "myregistry.com/library/nginx:1.19",
			expectedReg:    "myregistry.com",
			expectedName:   "library/nginx",
			expectedTag:    "1.19",
			expectedDigest: "",
			expectErr:      false,
		},
		{
			name:           "Valid image with digest",
			image:          "docker.io/library/nginx@sha256:abc123",
			expectedReg:    "docker.io",
			expectedName:   "library/nginx",
			expectedTag:    "",
			expectedDigest: "sha256:abc123",
			expectErr:      false,
		},
		{
			name:           "Invalid image format",
			image:          "/invalidimage",
			expectedReg:    "",
			expectedName:   "",
			expectedTag:    "",
			expectedDigest: "",
			expectErr:      true,
		},
		{
			name:           "Valid image with tag and digest",
			image:          "myregistry.com/library/nginx:1.19@sha256:abc123",
			expectedReg:    "myregistry.com",
			expectedName:   "library/nginx",
			expectedTag:    "1.19",
			expectedDigest: "sha256:abc123",
			expectErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg, name, tag, digest, err := ParseImage(tt.image)

			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedReg, reg)
				require.Equal(t, tt.expectedName, name)
				require.Equal(t, tt.expectedTag, tag)
				require.Equal(t, tt.expectedDigest, digest)
			}
		})
	}
}
