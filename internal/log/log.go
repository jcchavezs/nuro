package log

import (
	"io"
	"net/http"

	prettyconsole "github.com/thessem/zap-prettyconsole"
	"github.com/yuseferi/zax/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger = zap.NewNop()

func Init(loglevel zapcore.Level, output io.Writer) {
	Logger = prettyconsole.NewLogger(loglevel).
		WithOptions(zap.ErrorOutput(zapcore.AddSync(output)))
}

func Close() error {
	_ = Logger.Sync()
	return nil
}

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
