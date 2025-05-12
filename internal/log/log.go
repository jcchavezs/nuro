package log

import (
	"net/http"

	"github.com/yuseferi/zax/v2"
	"go.uber.org/zap"
)

var Logger *zap.Logger = zap.NewNop()

type logRoundTripper struct {
	http.RoundTripper
}

func (rt logRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	fields := []zap.Field{zap.String("method", req.Method)}
	fields = append(fields, zap.String("url", req.URL.String()))
	fields = append(fields, zax.Get(req.Context())...)

	res, err := rt.RoundTripper.RoundTrip(req)
	if err == nil {
		fields = append(fields, zap.Int("status", res.StatusCode))
	} else {
		fields = append(fields, zap.Error(err))
	}

	Logger.Debug("HTTP request", fields...)
	return res, err
}

func WrapRoundTripper(t http.RoundTripper) http.RoundTripper {
	return logRoundTripper{t}
}
