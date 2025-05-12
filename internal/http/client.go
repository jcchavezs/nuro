package http

import (
	"net/http"

	"github.com/jcchavezs/nuro/internal/auth"
	"github.com/jcchavezs/nuro/internal/log"
)

var Client = &http.Client{
	Transport: log.WrapRoundTripper(
		auth.WrapRoundTripper(http.DefaultTransport),
	),
}

var NewRequestWithContext = http.NewRequestWithContext

const StatusOK = http.StatusOK
