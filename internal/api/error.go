package api

import "errors"

// ErrorResponse represents an error response from a OCI server
type ErrorResponse struct {
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func (r ErrorResponse) Error() error {
	if len(r.Errors) == 0 {
		return nil
	}

	errs := make([]error, 0, len(r.Errors))
	for _, e := range r.Errors {
		errs = append(errs, errors.New(e.Message))
	}

	return errors.Join(errs...)
}
