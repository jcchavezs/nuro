package api

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestErrorResponse_Error(t *testing.T) {
	t.Run("no errors", func(t *testing.T) {
		resp := ErrorResponse{}
		err := resp.Error()
		require.NoError(t, err)
	})

	t.Run("single error", func(t *testing.T) {
		resp := ErrorResponse{
			Errors: []struct {
				Message string `json:"message"`
			}{
				{Message: "unauthorized access"},
			},
		}

		err := resp.Error()
		require.Error(t, err)
		require.Equal(t, "unauthorized access", err.Error())
	})

	t.Run("multiple errors", func(t *testing.T) {
		resp := ErrorResponse{
			Errors: []struct {
				Message string `json:"message"`
			}{
				{Message: "unauthorized access"},
				{Message: "invalid token"},
				{Message: "expired credentials"},
			},
		}

		err := resp.Error()
		require.Error(t, err)
		require.Contains(t, err.Error(), "unauthorized access")
		require.Contains(t, err.Error(), "invalid token")
		require.Contains(t, err.Error(), "expired credentials")
	})
}
