package middlewares

import (
	"context"
	"fmt"
	"github.com/farmani/service/foundation/web"
	"go.uber.org/zap"
	"net/http"
	"time"
)

// Logger writes information about the request to the logs.
func Logger(log *zap.SugaredLogger) web.Middleware {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			v := web.GetValues(ctx)

			path := r.URL.Path
			if r.URL.RawQuery != "" {
				path = fmt.Sprintf("%s?%s", path, r.URL.RawQuery)
			}

			log.Infow("request started", "trace_id", v.TraceID, "method", r.Method, "path", path, "remote_addr", r.RemoteAddr)

			err := handler(ctx, w, r)

			log.Infow("request completed", "trace_id", v.TraceID, "statuscode", v.StatusCode, "since", time.Since(v.Now), "method", r.Method, "path", path, "remote_addr", r.RemoteAddr)

			return err
		}

		return h
	}

	return m
}
